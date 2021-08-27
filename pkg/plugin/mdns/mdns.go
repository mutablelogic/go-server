package main

import (
	"context"
	"fmt"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/mdns"
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
		cfg.ServiceDatabase = []interface{}{mdns.DefaultServiceDatabase}
	}

	// Create mDNS server
	if server, err := mdns.New(cfg); err != nil {
		provider.Print(ctx, "NewServer: ", err)
		return nil
	} else {
		this.Server = server
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
	// Add handlers
	if err := this.AddHandlers(ctx, provider); err != nil {
		return err
	}

	// Enumerate services and instances in the background
	go func(parent context.Context) {
		time.Sleep(1 * time.Second)
		ctx, cancel := context.WithTimeout(parent, time.Second*10)
		defer cancel()
		services, err := this.Server.EnumerateServices(ctx)
		if err != nil {
			provider.Print(ctx, "EnumerateServices: ", err)
			return
		}
		ctx2, cancel2 := context.WithTimeout(parent, time.Second*10)
		defer cancel2()
		instances, err := this.Server.EnumerateInstances(ctx2, services...)
		if err != nil {
			provider.Print(ctx, "EnumerateInstances: ", err)
			return
		}
		for _, instance := range instances {
			provider.Print(ctx, "EnumerateInstances: ", instance.Instance())
		}
	}(ctx)

	// Run mDNS server
	return this.Server.Run(ctx)
}
