package main

import (
	"context"
	"fmt"
	"io"

	// Package imports
	highlight "github.com/zyedidia/highlight"

	// Namespace Imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	FileTypes []string `yaml:"types"`
}

type plugin struct {
	defs []*highlight.Def
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the text-renderer module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Read configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Printf(ctx, "Error reading config: %s", err)
		return nil
	}

	// Read definitions
	defs, err := highlight.AllDefs(cfg.FileTypes...)
	if err != nil {
		provider.Printf(ctx, "Failed to load highlight definitions: %s", err)
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
	for _, def := range p.defs {
		str += " " + def.FileType
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Renders text documents.\n")
	fmt.Fprintln(w, "  Configuration:")
	fmt.Fprintln(w, "    types: <list of string>")
	fmt.Fprintln(w, "      Required, the type of documents which should be syntax highlighted")
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
