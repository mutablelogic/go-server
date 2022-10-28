package logger

import (
	"context"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "log"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(_ context.Context, _ iface.Provider) (iface.Task, error) {
	if !types.IsIdentifier(p.Name()) {
		return nil, ErrBadParameter.With("Invalid plugin name: %q", p.Name())
	}
	if !types.IsIdentifier(p.Label()) {
		return nil, ErrBadParameter.With("Invalid plugin label: %q", p.Label())
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
