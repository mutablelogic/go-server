package task

import (
	// Package imports
	"context"
	"fmt"

	iface "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	event "github.com/mutablelogic/go-server/pkg/event"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugins is a map of all registered plugins
type provider struct {
	event.Source

	// Enumeration of tasks, keyed by name.label
	tasks map[string]iface.Task
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register adds a plugin to the map of plugins. It will return errors if the
// name or label is invalid, or the plugin with the same name already exists.
func NewProvider(parent context.Context, plugins ...iface.Plugin) (iface.Provider, error) {
	this := new(provider)
	this.tasks = make(map[string]iface.Task, len(plugins))
	plugins_ := Plugins{}
	if err := plugins_.Register(plugins...); err != nil {
		return nil, err
	}

	// TODO: Re-order the plugins so that dependencies are satisfied

	// Create the tasks sequentially, and return if any of the tasks
	// returns an error
	for _, plugin := range plugins {
		key := plugin.Name()
		parent := ctx.WithNameLabel(parent, plugin.Name(), plugin.Label())
		if task, err := plugin.New(parent, this); err != nil {
			return nil, err
		} else {
			this.tasks[key] = task
		}
	}

	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *provider) String() string {
	str := "<provider"
	for key, task := range p.tasks {
		str += fmt.Sprint(" ", key, "=", task)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *provider) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
