package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/httprouter"
	prv "github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Endpoint struct {
	// Prefix for the endpoint
	Prefix string `json:"prefix"`
	// Name of static serving which can be empty
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Filepath to serve static content
	Path string `yaml:"path" json:"-"`
}

type Config map[string]string

type plugin struct {
	alias []Endpoint
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Root static file serving from a directory
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

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
	for prefix, path := range cfg {
		if abspath, err := filepath.Abs(path); err != nil {
			provider.Print(ctx, err)
			return nil
		} else if stat, err := os.Stat(abspath); err != nil {
			provider.Print(ctx, err)
			return nil
		} else if !stat.IsDir() {
			provider.Print(ctx, "Not a folder: ", abspath)
			return nil
		} else {
			this.alias = append(this.alias, Endpoint{prefix, prefix, abspath})
		}
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<static"
	for _, endpoint := range this.alias {
		str += fmt.Sprintf(" %q => %q", endpoint.Prefix, endpoint.Path)
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
	for _, endpoint := range this.alias {
		if err := provider.AddHandler(prv.ContextWithPrefix(ctx, endpoint.Prefix), http.FileServer(http.Dir(endpoint.Path))); err != nil {
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

func (p *plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.ServeJSON(w, p.alias, http.StatusOK, 2)
}
