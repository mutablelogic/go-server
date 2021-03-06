package main

import (
	"context"
	"fmt"
	"io"

	// Namespace Imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the text-renderer module
func New(ctx context.Context, provider Provider) Plugin {
	return new(plugin)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<markdown-renderer"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Renders markdown documents.")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "markdown-renderer"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	<-ctx.Done()
	return nil
}
