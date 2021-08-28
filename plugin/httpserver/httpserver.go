package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	fcgi "github.com/djthorpe/go-server/pkg/fcgi"
	httprouter "github.com/djthorpe/go-server/pkg/httprouter"
	pr "github.com/djthorpe/go-server/pkg/provider"
	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Addr    string        `yaml:"addr"`    // Address or path for binding HTTP server
	Timeout time.Duration `yaml:"timeout"` // Read timeout on HTTP requests
	Key     string        `yaml:"key"`     // Path to TLS Private Key
	Cert    string        `yaml:"cert"`    // Path to TLS Certificate
}

type server struct {
	Config
	Router
	srv  *http.Server
	fcgi *fcgi.Server
	log  Logger
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultTimeout = 10 * time.Second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(server)
	this.Router = httprouter.New()

	// Get logger
	this.log = provider.GetPlugin(ctx, "log").(Logger)

	// Get configuration
	if err := provider.GetConfig(ctx, &this.Config); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Override addr if set in context
	if addr := pr.ContextAddr(ctx); addr != "" {
		this.Config.Addr = addr
	}

	// Check addr for being (host,port). If not, then run as FCGI server
	if _, _, err := net.SplitHostPort(this.Config.Addr); this.Config.Addr != "" && err != nil {
		if err := this.fcgiserver(this.Config.Addr); err != nil {
			provider.Print(ctx, err)
			return nil
		}
	}

	// If either key or cert is non-nil then create a TLSConfig
	var tlsconfig *tls.Config
	if this.Config.Key != "" || this.Config.Cert != "" {
		if cert, err := tls.LoadX509KeyPair(this.Config.Cert, this.Config.Key); err != nil {
			provider.Print(ctx, err)
			return nil
		} else {
			tlsconfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}
	}

	// If addr is empty, then set depending on whether it's SSL or not
	if this.Config.Addr == "" {
		if tlsconfig == nil {
			this.Config.Addr = ":http"
		} else {
			this.Config.Addr = ":https"
		}
	}

	// Create net server
	if err := this.netserver(this.Config.Addr, tlsconfig, this.Config.Timeout); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *server) String() string {
	str := "<httpserver"
	if this.fcgi != nil {
		str += fmt.Sprintf(" fcgi=%q", this.fcgi.Addr)
	} else {
		str += fmt.Sprintf(" addr=%q", this.srv.Addr)
		if this.srv.TLSConfig != nil {
			str += " tls=true"
		}
		if this.srv.ReadHeaderTimeout != 0 {
			str += fmt.Sprintf(" read_timeout=%v", this.srv.ReadHeaderTimeout)
		}
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MODULE

func Name() string {
	return "httpserver"
}

func (this *server) Run(ctx context.Context, _ Provider) error {
	var result error
	go func() {
		<-ctx.Done()
		if err := this.stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}()
	if err := this.runInForeground(); err != nil && errors.Is(err, http.ErrServerClosed) == false {
		result = multierror.Append(result, err)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *server) fcgiserver(path string) error {
	// Create server
	this.fcgi = &fcgi.Server{}
	this.fcgi.Handler = this.Router.(http.Handler)
	this.fcgi.Network = "unix"
	this.fcgi.Addr = path

	// Return success
	return nil
}

func (this *server) netserver(addr string, config *tls.Config, timeout time.Duration) error {
	// Set up server
	this.srv = &http.Server{}
	if config != nil {
		this.srv.TLSConfig = config
	}

	// Set default timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	// Set server parameters
	this.srv.Addr = addr
	this.srv.Handler = this.Router.(http.Handler)
	this.srv.ReadHeaderTimeout = timeout
	this.srv.IdleTimeout = timeout

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// START AND STOP

func (this *server) runInForeground() error {
	if this.fcgi != nil {
		return this.fcgi.ListenAndServe()
	} else if this.srv.TLSConfig != nil {
		return this.srv.ListenAndServeTLS("", "")
	} else {
		return this.srv.ListenAndServe()
	}
}

func (this *server) stop() error {
	if this.fcgi != nil {
		return this.fcgi.Close()
	} else {
		return this.srv.Close()
	}
}
