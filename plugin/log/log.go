package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	// Modules
	"github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Prefix string   `yaml:"prefix"`
	Flags  []string `yaml:"flags"`
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
		p.Printf(ctx, "%s %s", r.Method, r.URL.Path)
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
	fmt.Fprintln(w, "  Logs output to stderr and includes debugging information.")
	fmt.Fprintln(w, "\n  prefix: <string>")
	fmt.Fprintln(w, "    Optional, include a prefix before each log entry")
	fmt.Fprintln(w, "\n  flags: <array of string>")
	fmt.Fprintln(w, "    Optional, can include 'default', 'date', 'time', 'microseconds', 'utc' and 'msgprefix'")
}
