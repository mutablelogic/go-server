package main

import (
	"context"
	"log"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type logger struct{}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	return new(logger)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *logger) String() string {
	str := "<log"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MODULE

func Name() string {
	return "log"
}

func (this *logger) Run(context.Context) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOG

func (this *logger) Print(_ context.Context, v ...interface{}) {
	log.Print(v...)
}

func (this *logger) Printf(_ context.Context, fmt string, v ...interface{}) {
	log.Printf(fmt, v...)
}
