package task

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	iface "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	event "github.com/mutablelogic/go-server/pkg/event"
	"github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"

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
	order []string
	tasks map[string]iface.Task
	log   []plugin.Log
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register adds a plugin to the map of plugins. It will return errors if the
// name or label is invalid, or the plugin with the same name already exists.
func NewProvider(parent context.Context, plugins ...iface.Plugin) (iface.Provider, error) {
	this := new(provider)
	this.order = make([]string, 0, len(plugins))
	this.tasks = make(map[string]iface.Task, len(plugins))
	plugins_ := Plugins{}
	if err := plugins_.Register(plugins...); err != nil {
		return nil, err
	}

	// Create a graph of the plugins
	// Re-order the plugins so that dependencies are satisfied correctly
	// The order returned is the order in which the plugins should be created
	order, err := NewGraph(plugins...).Resolve()
	if err != nil {
		return nil, err
	}

	// Create the tasks sequentially, and return if any error is returned
	for _, plugin := range order {
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

func (p *provider) Run(parent context.Context) error {
	var wg sync.WaitGroup
	var result error

	// Make cancelable, so that on any error all tasks are stopped
	child, cancel := context.WithCancel(parent)
	defer cancel()

	// Run all tasks
	for label, task := range p.tasks {
		// Add label to context - UGLY!
		nameLabel := strings.SplitN(label, ".", 2)
		if len(nameLabel) != 2 {
			result = multierror.Append(result, ErrBadParameter.Withf("Invalid label: %q", label))
			continue
		}

		// Subscribe to events from task
		wg.Add(1)
		go func(label string, task iface.Task) {
			defer wg.Done()
			if err := p.run(ctx.WithNameLabel(child, nameLabel[0], nameLabel[1]), task); err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					result = multierror.Append(result, fmt.Errorf("%v: %w", label, err))
				}
				cancel()
			}
		}(label, task)
	}

	// Wait until all tasks are completed
	wg.Wait()

	// Close the event source
	if err := p.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

// New creates a new task from a plugin. It should only be called from
// the 'new' function, not once the provider is in Run state.
func (p *provider) New(parent context.Context, proto iface.Plugin) (iface.Task, error) {
	key := KeyForPlugin(proto)
	if task := p.Get(key); task != nil {
		return nil, ErrDuplicateEntry.Withf("Duplicate task: %q", key)
	}

	// Resolve dependencies
	p.order = append(p.order, key)
	resolveRef(key, reflect.ValueOf(proto), func(ref string, v reflect.Value) error {
		// Get the task from the reference
		var task iface.Task
		if ref != "" {
			task = p.Get(ref)
			if task == nil {
				return ErrNotFound.Withf("%q: Task not found: %q", key, ref)
			}
		}

		// Assign task to plugin
		if t, ok := v.Interface().(types.Task); !ok {
			return ErrBadParameter.Withf("%q: Invalid type: %v", key, v.Type())
		} else {
			t.Task = task
		}

		// Return success
		return nil
	})

	// Create the task
	task, err := proto.New(ctx.WithNameLabel(parent, proto.Name(), proto.Label()), p)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", key, err)
	} else if err := p.Set(key, task); err != nil {
		return nil, fmt.Errorf("%v: %w", key, err)
	}

	// Check for log task
	if log, ok := task.(plugin.Log); ok && log != nil {
		p.log = append(p.log, log)
	}

	// Return success
	return task, nil
}

// Print log message to any log tasks registered
func (p *provider) Print(ctx context.Context, v ...any) {
	for _, log := range p.log {
		log.Print(ctx, v...)
	}
}

// Format and print log message
func (p *provider) Printf(ctx context.Context, format string, v ...any) {
	for _, log := range p.log {
		log.Printf(ctx, format, v...)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Keys returns the list of task keys in the provider, ordered as per the
// order in which the tasks were added.
func (p *provider) Keys() []string {
	p.RLock()
	defer p.RUnlock()
	result := make([]string, 0, len(p.tasks))
	for key := range p.tasks {
		result = append(result, key)
	}
	return result
}

func (p *provider) Get(args ...string) iface.Task {
	p.RLock()
	defer p.RUnlock()

	// Exact match
	key := strings.Trim(strings.Join(args, "."), ".")
	if task, exists := p.tasks[key]; exists {
		return task
	}
	// Prefix match
	for key2, task := range p.tasks {
		if strings.HasPrefix(key2, key+".") {
			return task
		}
	}
	// No match
	return nil
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

func (p *provider) run(child context.Context, task iface.Task) error {
	var wg sync.WaitGroup

	// Cancel on any error
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		ch := task.Sub()
		for {
			select {
			case <-ctx.Done():
				task.Unsub(ch)
				return
			case event := <-ch:
				p.Emit(event)
			}
		}
	}(ctx)

	// Start child, then cancel the event receiver - UGLY!
	err := task.Run(child)
	cancel()
	wg.Wait()
	return err
}
