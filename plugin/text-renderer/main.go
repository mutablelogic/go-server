package main

import (
	"context"

	// Package imports
	highlight "github.com/zyedidia/highlight"

	// Namespace Imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
	defs []*highlight.Def
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the text-renderer module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Read in all definitions
	defs, err := highlight.AddDefs()
	if err != nil {
		provider.Print(ctx, "Failed to load highlight definitions: %s", err)
		return nil
	} else {
		p.defs = defs
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<text-renderer"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "text-renderer"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	<-ctx.Done()
	return nil
}
