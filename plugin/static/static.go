package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	"github.com/mutablelogic/go-server/pkg/config"
	router "github.com/mutablelogic/go-server/pkg/httprouter"
	prv "github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Endpoint struct {
	// Prefix for the endpoint
	Prefix string `json:"prefix"`
	// Name of static serving which can be empty
	Name string `json:"name,omitempty"`
	// Handler for serving content, including middleware
	Handler http.Handler `json:"-"`
	// Middleware
	Middleware []string `json:"-"`
}

type plugin struct {
	sync.Mutex
	handlers []Endpoint
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	pathSeparator = string(os.PathSeparator)
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Root static file serving from a directory
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)
	p.handlers = make([]Endpoint, 0)

	// Return success
	return p
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	var result error

	// Add handlers for static content
	for _, endpoint := range p.handlers {
		child := prv.ContextWithHandler(ctx, config.Handler{
			Prefix:     endpoint.Prefix,
			Middleware: endpoint.Middleware,
		})
		if err := provider.AddHandler(child, endpoint.Handler); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Add handler for API
	if err := provider.AddHandler(ctx, p); err != nil {
		result = multierror.Append(result, err)
	}

	// Quit if any errors
	if result != nil {
		return result
	}

	// Wait for shutdown
	<-ctx.Done()

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<static"
	for _, endpoint := range p.handlers {
		str += fmt.Sprintf(" %v=%q", endpoint.Name, endpoint.Prefix)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "static"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - STATIC

func (p *plugin) AddFS(ctx context.Context, fs fs.FS, suffix string) error {
	p.Lock()
	defer p.Unlock()

	name := prv.ContextPluginName(ctx)
	if name == "" {
		return ErrBadParameter.With("Missing plugin name")
	}
	prefix := prv.ContextHandlerPrefix(ctx)
	if prefix == "" {
		return ErrBadParameter.With("Missing prefix")
	}
	p.handlers = append(p.handlers, Endpoint{
		Prefix:     pathSeparator + strings.TrimPrefix(prefix, pathSeparator),
		Name:       name,
		Handler:    NewFileSystemHandler(fs, suffix),
		Middleware: prv.ContextHandlerMiddleware(ctx),
	})

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - HANDLER

func (p *plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.ServeJSON(w, p.handlers, http.StatusOK, 2)
}
