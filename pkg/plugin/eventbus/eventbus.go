package main

import (
	"context"
	"fmt"

	// Modules
	. "github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the eventbus module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<eventbus"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "eventbus"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	// Wait until done
	<-ctx.Done()

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - EVENT BUS

func (this *plugin) Post(ctx context.Context, evt Event) {
	fmt.Println(provider.DumpContext(ctx), evt)
}

func (this *plugin) Subscribe(ctx context.Context, _ chan<- Event) {
	fmt.Println(provider.DumpContext(ctx), "TODO: Subscribe")
}
