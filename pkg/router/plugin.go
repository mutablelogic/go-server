package router

import (
	"context"
	"os"

	// Namespace imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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
	// Check name and label
	if !types.IsIdentifier(p.Name()) {
		return nil, ErrBadParameter.Withf("Invalid plugin name: %q", p.Name())
	}
	if !types.IsIdentifier(p.Label()) {
		return nil, ErrBadParameter.Withf("Invalid plugin label: %q", p.Label())
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
