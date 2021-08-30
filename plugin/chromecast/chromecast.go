package main

import (
	"context"
	"sync"

	// Modules
	. "github.com/djthorpe/go-server"
	chromecast "github.com/djthorpe/go-server/pkg/chromecast"
	"github.com/djthorpe/go-server/pkg/mdns"
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
	*chromecast.Manager
	C chan Event
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 100
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the chomecast module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)
	this.C = make(chan Event, defaultCapacity)

	// Load configuration
	cfg := chromecast.Config{}
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	} else {
		this.Manager = chromecast.NewManager(cfg)
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<chromecast"
	if this.Manager != nil {
		str += " " + this.Manager.String()
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "chromecast"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	var wg sync.WaitGroup
	var result error

	// Subscribe to mDNS events
	if err := provider.Subscribe(ctx, this.C); err != nil {
		return err
	}

	// Run chromecast manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.Manager.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Receive events from mDNS
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.ReceiveEvents(ctx, provider); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Wait for all go routines to end
	wg.Wait()

	// Close channel
	close(this.C)

	// Return success
	return nil
}

func (this *plugin) ReceiveEvents(ctx context.Context, _ Provider) error {
	// Process events until done
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-this.C:
			if evt := filterEvent(evt); evt != nil {
				switch evt.Type() {
				case mdns.EVENT_TYPE_ADDED, mdns.EVENT_TYPE_CHANGED:
					this.Add(evt.Service())
				case mdns.EVENT_TYPE_EXPIRED, mdns.EVENT_TYPE_REMOVED:
					this.Remove(evt.Service())
				}
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func filterEvent(evt Event) mdns.ServiceEvent {
	if evt, ok := evt.(mdns.ServiceEvent); ok && evt.Service().Service() == chromecast.ServiceName {
		return evt
	} else {
		return nil
	}
}
