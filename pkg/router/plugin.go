package router

import (
	"context"
	"os"

	// Namespace imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Plugin for the router does not currently contain any tunable values
type Plugin struct {
	task.Plugin
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName   = "router"
	pathSeparator = string(os.PathSeparator)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(_ context.Context, _ iface.Provider) (iface.Task, error) {
	// Check parameters
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	return NewWithPlugin(p)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}
