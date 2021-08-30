package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	// Modules
	. "github.com/djthorpe/go-server"
	router "github.com/djthorpe/go-server/pkg/httprouter"
	prv "github.com/djthorpe/go-server/pkg/provider"
	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Endpoint struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type Config map[string]Endpoint

type plugin struct {
	alias map[string]Endpoint
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Root static file serving from a directory
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)
	this.alias = make(map[string]Endpoint)

	// Load configuration
	cfg := Config{}
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	} else if len(cfg) == 0 {
		provider.Print(ctx, `Requires "path" argument`)
		return nil
	}

	// Make paths absolute
	for prefix, endpoint := range cfg {
		if abspath, err := filepath.Abs(endpoint.Path); err != nil {
			provider.Print(ctx, err)
			return nil
		} else if stat, err := os.Stat(abspath); err != nil {
			provider.Print(ctx, err)
			return nil
		} else if !stat.IsDir() {
			provider.Print(ctx, "Not a folder: ", endpoint.Path)
			return nil
		} else {
			this.alias[prefix] = Endpoint{endpoint.Name, endpoint.Path}
		}
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<static"
	for alias, endpoint := range this.alias {
		str += fmt.Sprintf(" %q =>  %q", alias, endpoint)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PULGIN

func Name() string {
	return "static"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	var result error

	// Add handlers for static content
	for alias, endpoint := range this.alias {
		if err := provider.AddHandler(prv.ContextWithPrefix(ctx, alias), http.FileServer(http.Dir(endpoint.Path))); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Add handler for API
	if err := provider.AddHandler(ctx, this); err != nil {
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
// PUBLIC METHODS - HANDLER

func (this *plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.ServeJSON(w, this.alias, http.StatusOK, 2)
}
