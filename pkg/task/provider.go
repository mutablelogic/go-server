package task

import (
	"context"
	"fmt"
	"sync"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	event "github.com/mutablelogic/go-server/pkg/event"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugins is a map of all registered plugins
type provider struct {
	event.Source
	sync.RWMutex

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

	// Create the tasks sequentially, and return if any error is returned
	for _, plugin := range plugins {
		if _, err := this.New(parent, plugin); err != nil {
			return nil, err
		}
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *provider) String() string {
	str := "<provider"
	for _, key := range p.Keys() {
		str += fmt.Sprint(" ", key, "=", p.Get(key))
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *provider) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// New creates a new task from a plugin. It should only be called from
// the 'new' function, not once the provider is in Run state.
func (p *provider) New(parent context.Context, plugin iface.Plugin) (iface.Task, error) {
	key := plugin.Name() + "." + plugin.Label()
	if task := p.Get(key); task != nil {
		return nil, ErrDuplicateEntry.Withf("Duplicate task: %q", key)
	} else if task, err := plugin.New(ctx.WithNameLabel(parent, plugin.Name(), plugin.Label()), p); err != nil {
		return nil, err
	} else if err := p.Set(key, task); err != nil {
		return nil, err
	} else {
		return task, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *provider) Keys() []string {
	p.RLock()
	defer p.RUnlock()
	result := make([]string, 0, len(p.tasks))
	for key := range p.tasks {
		result = append(result, key)
	}
	return result
}

func (p *provider) Get(key string) iface.Task {
	p.RLock()
	defer p.RUnlock()
	return p.tasks[key]
}

func (p *provider) Set(key string, task iface.Task) error {
	p.Lock()
	defer p.Unlock()
	if _, exists := p.tasks[key]; exists {
		return ErrDuplicateEntry.Withf("Duplicate task: %q", key)
	}
	p.tasks[key] = task
	return nil
}
