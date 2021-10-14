package main

import (
	"context"
	"fmt"
	"io"

	// Package imports

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
	p := new(plugin)

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<zip-renderer"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Renders .ZIP and .TAR files.\n")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "zip-renderer"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	<-ctx.Done()
	return nil
}
