package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	// Modules
	"github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/mdns"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type server struct {
	*mdns.Server
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the mdns module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(server)

	// Load configuration
	var cfg mdns.Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Set default service database if not set
	if len(cfg.ServiceDatabase) == 0 {
		cfg.ServiceDatabase = []string{mdns.DefaultServiceDatabase}
	}

	// Create mDNS server
	if server, err := mdns.New(cfg); err != nil {
		provider.Print(ctx, "NewServer: ", err)
		return nil
	} else {
		this.Server = server
		this.Server.C = make(chan Event, 100)
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *server) String() string {
	str := "<mdns"
	if this.Server != nil {
		str += " " + fmt.Sprint(this.Server)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "mdns"
}

func (this *server) Run(ctx context.Context, provider Provider) error {
	var wg sync.WaitGroup
	var result error

	// Add handlers
	if err := this.AddHandlers(ctx, provider); err != nil {
		return err
	}

	// Enumerate services and instances in the background
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		// Enumerate services and instances after a startup delay
		timer := time.NewTimer(time.Second * 5)
		defer timer.Stop()

		select {
		case <-timer.C:
			ctx2, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			services, err := this.Server.EnumerateServices(ctx2)
			if err != nil {
				provider.Print(ctx, "EnumerateServices: ", err)
				return
			}
			ctx3, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if _, err = this.Server.EnumerateInstances(ctx3, services...); err != nil {
				provider.Print(ctx, "EnumerateInstances: ", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}(ctx)

	// Send mDNS events onto the event bus
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case event := <-this.Server.C:
				if event != nil {
					provider.Post(ctx, event)
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	// Run mDNS server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.Server.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Close event channel
	close(this.Server.C)

	// Run any errors
	return result
}
