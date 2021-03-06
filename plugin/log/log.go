package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	// Packages
	provider "github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Label  string   `hcl:",label"`
	Prefix string   `yaml:"prefix" hcl:"prefix"`
	Flags  []string `yaml:"flags" hcl:"flags,optional"`
}

type plugin struct {
	sync.Mutex
	*log.Logger
}

type handler struct {
	http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Get Configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Get flags
	flags, err := flagsForSlice(cfg.Flags)
	if err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Make logger
	if log := log.New(os.Stderr, cfg.Prefix, flags); log == nil {
		provider.Print(ctx, "Cannot create logger")
		return nil
	} else {
		p.Logger = log
	}

	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<log"
	if prefix := p.Logger.Prefix(); prefix != "" {
		str += fmt.Sprintf(" prefix=%q", prefix)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "log"
}

func GetConfig() interface{} {
	return &Config{}
}

func (p *plugin) Run(ctx context.Context, _ Provider) error {
	<-ctx.Done()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOG

func (p *plugin) Print(ctx context.Context, v ...interface{}) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if name := provider.ContextPluginName(ctx); name != "" {
		p.Logger.Print("["+name+"] ", fmt.Sprint(v...))
	} else {
		p.Logger.Print(v...)
	}
}

func (p *plugin) Printf(ctx context.Context, f string, v ...interface{}) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if name := provider.ContextPluginName(ctx); name != "" {
		p.Logger.Print("["+name+"] ", fmt.Sprintf(f, v...))
	} else {
		p.Logger.Printf(f, v...)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (p *plugin) AddMiddlewareFunc(ctx context.Context, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlpath := r.URL.Path
		if prefix := provider.ContextHandlerPrefix(ctx); prefix != "" {
			urlpath = path.Join(prefix, urlpath)
		}
		p.Printf(ctx, "%s %s", r.Method, urlpath)
		h(w, r)
	}
}

func (p *plugin) AddMiddleware(ctx context.Context, h http.Handler) http.Handler {
	return &handler{
		p.AddMiddlewareFunc(ctx, h.ServeHTTP),
	}
}

func (p *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.HandlerFunc(w, r)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - MIDDLEWARE

func flagsForSlice(flags []string) (int, error) {
	var result int
	for _, flag := range flags {
		flag = strings.ToLower(flag)
		switch flag {
		case "default", "standard", "std":
			result |= log.LstdFlags
		case "date":
			result |= log.Ldate
		case "time":
			result |= log.Ltime
		case "microseconds", "ms":
			result |= log.Lmicroseconds
		case "utc":
			result |= log.LUTC
		case "msgprefix", "prefix":
			result |= log.Lmsgprefix
		default:
			return 0, ErrBadParameter.Withf("flag: %q", flag)
		}
	}
	// Return success
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Logs output to stderr and includes debugging information.\n")
	fmt.Fprintln(w, "  Configuration:")
	fmt.Fprintln(w, "    prefix: <string>")
	fmt.Fprintln(w, "      Optional, include a prefix before each log entry")
	fmt.Fprintln(w, "    flags: <list of string>")
	fmt.Fprintln(w, "      Optional, can include 'default', 'date', 'time', 'microseconds', 'utc' and 'msgprefix'")
}
