package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
	"unicode"

	// Packages
	"github.com/mutablelogic/go-server/pkg/ddregister"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Timeout time.Duration `yaml:"timeout"`
	Delta   time.Duration `yaml:"delta"`
	Hosts   []string      `yaml:"hosts"`
}

type plugin struct {
	Config
	*ddregister.Registration

	modtime map[string]time.Time // Last time each hostname was registered
	addr    net.IP               // Current external address
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	minDelta     = 5 * time.Minute
	defaultDelta = 30 * time.Minute
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)
	p.modtime = make(map[string]time.Time)

	// Load configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	} else {
		p.Config = cfg
	}

	// Create registration
	if cfg.Timeout != 0 {
		p.Registration = ddregister.NewWithConfig(ddregister.Config{
			Timeout: cfg.Timeout,
		})
	} else {
		p.Registration = ddregister.New()
	}

	// Set default request delta
	if p.Delta == 0 {
		p.Delta = defaultDelta
	} else {
		p.Delta = maxDuration(p.Delta, minDelta)
	}

	// Generate external address
	if addr, err := p.GetExternalAddress(); err != nil {
		provider.Print(ctx, "GetExternalAddress: ", err)
		return nil
	} else {
		p.modtime["*"] = time.Now()
		p.addr = addr
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<ddregister"
	if this.addr != nil {
		str += fmt.Sprint(" addr=", this.addr)
	}
	if this.Delta != 0 {
		str += fmt.Sprint(" delta=", this.Delta)
	}
	str += " hosts=["
	for _, host := range this.Hosts {
		if host, _, _ := this.getcredentials(host); host != "" {
			str += fmt.Sprint(" ", host)
		}
	}
	str += " ]"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "ddregister"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	// Create ticker for doing the registration
	ticker := time.NewTicker(31 * time.Second)
	defer ticker.Stop()

	// Run loop
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get external address
			if p.touch("*") {
				if addr, err := p.GetExternalAddress(); err != nil {
					p.addr = nil
					provider.Print(ctx, "GetExternalAddress: ", err)
				} else if p.addr.String() != addr.String() {
					provider.Print(ctx, "GetExternalAddress: Changed to ", addr)
					p.addr = addr
					// Re-register all hosts
					for k := range p.modtime {
						if k != "*" {
							delete(p.modtime, k)
						}
					}
				}
			} else if host, user, passwd := p.next(); host != "" {
				// Register address - may return nil or ErrNotModified on success
				if err := p.RegisterAddress(host, user, passwd, p.addr, false); err != nil && !errors.Is(err, ErrNotModified) {
					provider.Printf(ctx, "RegisterAddress: %q: %v", host, err)
				}
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return an entry to register
func (p *plugin) next() (string, string, string) {
	for _, host := range p.Hosts {
		if host, user, passwd := p.getcredentials(host); host != "" {
			if p.touch(host) && p.addr != nil {
				return host, user, passwd
			}
		}
	}
	// No entry found
	return "", "", ""
}

// touch the modtime for the given hostname to see if we should re-register
func (p *plugin) touch(key string) bool {
	modtime, exists := p.modtime[key]
	if !exists || modtime.IsZero() {
		p.modtime[key] = time.Now()
		return true
	}
	if time.Since(modtime) >= p.Delta {
		p.modtime[key] = time.Now()
		return true
	}
	return false
}

// split out credentials from host entry
func (p *plugin) getcredentials(entry string) (string, string, string) {
	creds := strings.FieldsFunc(entry, func(r rune) bool {
		return r == ':' || unicode.IsSpace(r)
	})
	if len(creds) != 3 {
		return "", "", ""
	} else {
		return strings.ToLower(creds[0]), creds[1], creds[2]
	}
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
