package router

import (
	"context"
	"os"

	// Packages
	iface "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Plugin for the router maps prefixes to gateways
type Plugin struct {
	task.Plugin
	Prefix_     types.String `json:"prefix,omitempty"`     // Path for serving the router schema, optional
	Routes      []Route      `json:"routes"`               // Routes to add to the router, optional (but useless without)
	Middleware_ []string     `json:"middleware,omitempty"` // Middleware to add to the router for all routes, optional
}

type Route struct {
	Prefix      string     `json:"prefix"`               // Prefix path for the gateway service
	Handler     types.Task `json:"service"`              // Service handler
	Middleware_ []string   `json:"middleware,omitempty"` // Middleware to add to the router for this route, optional
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName   = "httpserver-router"
	pathSeparator = string(os.PathSeparator)
	hostSeparator = "."
)

const (
	ScopeRead = "github.com/mutablelogic/go-server/router:read"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(parent context.Context, provider iface.Provider) (iface.Task, error) {
	// Check parameters
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Return router
	return NewWithPlugin(p, ctx.NameLabel(parent))
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(defaultName, label),
	}
}

func (p Plugin) WithPrefix(prefix string) Plugin {
	p.Prefix_ = types.String(prefix)
	return p
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

func (p Plugin) Prefix() string {
	return string(p.Prefix_)
}

func (p Plugin) Middleware() []string {
	return p.Middleware_
}
