package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type static struct {
	Path string `yaml:"path"`
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Root static file serving from a directory
func New(ctx context.Context, provider Provider) Plugin {
	this := new(static)

	// Load configuration
	if err := provider.GetConfig(ctx, this); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	} else if this.Path == "" {
		provider.Print(ctx, `Requires "path" argument`)
		return nil
	}

	// Make path absolute
	if abspath, err := filepath.Abs(this.Path); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		this.Path = abspath
	}

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

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *static) String() string {
	str := "<static"
	if this.Path != "" {
		str += fmt.Sprintf(" path=%q", this.Path)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MODULE

func Name() string {
	return "static"
}

func (this *static) Run(context.Context) error {
	return nil
}
