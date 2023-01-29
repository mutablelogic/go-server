package main

import (
	"context"

	// Packages
	auth "github.com/mutablelogic/go-accessory/pkg/auth"
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Pool_ types.Task `json:"pool"` // Connection Pool
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName  = "accessory-tokenauth"
	defaultAdmin = "root"
)

var (
	// Default admin scopes
	adminScopes = []string{auth.ScopeRead, auth.ScopeWrite}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new connection pool task from plugin configuration
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	} else {
		return NewWithPlugin(p)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a plugin with a specific label
func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(defaultName, label),
	}
}

// Return the name of the plugin
func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

// Return the Connection Pool for the plugin
func (p Plugin) Pool() plugin.ConnectionPool {
	if pool, ok := p.Pool_.Task.(plugin.ConnectionPool); ok {
		return pool
	} else {
		return nil
	}
}
