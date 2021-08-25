package main

import (
	"context"
	"sync"

	// Modules
	. "github.com/djthorpe/go-server"
	rotel "github.com/djthorpe/go-server/pkg/rotel"
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type server struct {
	sync.Mutex
	*rotel.Manager

	c chan rotel.Event
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(server)
	this.c = make(chan rotel.Event)
	if manager, err := rotel.New(rotel.Config{}, this.c); err != nil {
		provider.Print(ctx, err.Error())
		return nil
	} else {
		this.Manager = manager
	}
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *server) String() string {
	return this.Manager.String()
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "rotel"
}

func (this *server) Run(ctx context.Context, provider Provider) error {
	var wg sync.WaitGroup
	var result error

	// Run manager in background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.Manager.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Run events
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.RunEvents(ctx, provider); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Release resources
	close(this.c)

	// Return any errors
	return result
}

func (this *server) RunEvents(ctx context.Context, provider Provider) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-this.c:
			provider.Print(ctx, evt)
		}
	}
}
