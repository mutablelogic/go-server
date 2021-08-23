package main

import (
	"context"
	"log"
	"sync"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type logger struct {
	sync.Mutex
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(logger)
	return this
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
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	log.Print(v...)
}

func (this *logger) Printf(_ context.Context, fmt string, v ...interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	log.Printf(fmt, v...)
}
