package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config map[string]string

type plugin struct {
	alias map[string]string
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Root static file serving from a directory
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)
	this.alias = make(map[string]string)

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
			provider.Print(ctx, "Not a folder: ", path)
			return nil
		} else {
			this.alias[prefix] = abspath
		}
	}
	/*
		// Check path, add handler
		if stat, err := os.Stat(this.Path); err != nil {
			provider.Print(ctx, err)
			return nil
		} else if !stat.IsDir() {
			provider.Print(ctx, "Not a folder: ", this.Path)
			return nil
		} else if err := provider.AddHandler(ctx, http.FileServer(http.Dir(this.Path))); err != nil {
			provider.Print(ctx, "Failed to add handler: ", err)
			return nil
		}
	*/

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<static"
	for alias, path := range this.alias {
		str += fmt.Sprintf(" %q =>  %q", alias, path)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MODULE

func Name() string {
	return "static"
}

func (this *plugin) Run(ctx context.Context, _ Provider) error {
	<-ctx.Done()
	return nil
}
