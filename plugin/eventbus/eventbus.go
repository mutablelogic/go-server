package main

import (
	"context"
	"fmt"

	// Modules
	. "github.com/djthorpe/go-server"
	provider "github.com/djthorpe/go-server/pkg/provider"
	sq "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Database string `yaml:"database"`
}

type plugin struct {
	sq.SQConnection
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultDatabase = "main"
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the eventbus module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Get sqlite
	if conn, ok := provider.GetPlugin(ctx, "sqlite").(sq.SQConnection); !ok {
		provider.Print(ctx, "missing sqlite dependency")
		return nil
	} else {
		this.SQConnection = conn
	}

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
	if err := this.createTables(ctx); err != nil {
		provider.Print(ctx, "failed to create tables: ", err)
		return err
	}

	// Wait until done
	<-ctx.Done()

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - EVENT BUS

func (this *plugin) Post(ctx context.Context, evt Event) {
	fmt.Println(provider.DumpContext(ctx), evt)
	go func() {
		this.indexEvent(ctx, evt)
	}()
}

func (this *plugin) Subscribe(ctx context.Context, _ chan<- Event) {
	fmt.Println(provider.DumpContext(ctx), "TODO: Subscribe")
}
