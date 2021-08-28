package main

import (
	"context"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the mdns module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<chromecast"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "chromecast"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	// Wait until done
	<-ctx.Done()

	// Return success
	return nil
}
