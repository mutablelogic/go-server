package main

import (
	"context"
	"fmt"
	"io"

	// Packages
	frontend "github.com/mutablelogic/go-server/npm/events"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct{}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the template module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Setup static serving
	if static, ok := provider.GetPlugin(ctx, "static").(Static); !ok {
		provider.Print(ctx, "no static plugin")
		return nil
	} else if err := static.AddFS(ctx, frontend.Dist, "dist"); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	return "<events-frontend>"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Static files for HTML frontend for events.")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "events-frontend"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	<-ctx.Done()
	return nil
}
