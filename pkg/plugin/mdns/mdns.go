package main

import (
	"context"
	"fmt"

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

	// Create mDNS server
	if server, err := mdns.New(cfg); err != nil {
		provider.Print(ctx, "NewServer: ", err)
		return nil
	} else {
		this.Server = server
	}

	// Add handler for instances
	if err := provider.AddHandlerFuncEx(ctx, reRouteInstances, this.ServeInstances); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
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

func (this *server) Run(ctx context.Context, _ Provider) error {
	return this.Server.Run(ctx)
}
