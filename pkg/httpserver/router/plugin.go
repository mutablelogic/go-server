package router

import (
	"context"
	"os"

	// Packages
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Plugin for the router maps prefixes to gateways
type Plugin struct {
	task.Plugin
	Routes []Route `json:"routes"`
}

type Route struct {
	Prefix  string     `json:"prefix"`
	Handler types.Task `json:"handler"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName   = "router"
	pathSeparator = string(os.PathSeparator)
	hostSeparator = "."
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(_ context.Context, _ iface.Provider) (iface.Task, error) {
	// Check parameters
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Check gateway handlers are of type iface.Gateway
	gateways := make(map[string]plugin.Gateway, len(p.Routes))
	for _, gateway := range p.Routes {
		route := NewRoute(gateway.Prefix, nil, nil).Prefix()
		if handler, ok := gateway.Handler.Task.(plugin.Gateway); !ok {
			return nil, ErrBadParameter.Withf("Handler for %q is not a gateway", gateway.Prefix)
		} else if _, exists := gateways[route]; exists {
			return nil, ErrDuplicateEntry.Withf("Duplicate prefix %q", gateway.Prefix)
		} else {
			gateways[route] = handler
		}
	}

	return NewWithPlugin(p, gateways)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(defaultName, label),
	}
}

func (p Plugin) WithRoutes(r []Route) Plugin {
	p.Routes = r
	return p
}

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}
