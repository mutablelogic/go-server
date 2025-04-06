package provider

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type provider struct {
	// Map labels to plugins
	plugin map[string]server.Plugin

	// Map labels to tasks
	task map[string]*state

	// Order that the tasks were created
	order []string

	// Function to resolve plugin members
	resolver ResolverFunc

	// Default logger
	server.Logger `json:"-"`
}

var _ server.Provider = (*provider)(nil)

type state struct {
	server.Task
	context.Context
	context.CancelFunc
	sync.WaitGroup
}

// ResolverFunc is a function that resolves a plugin from label and plugin
type ResolverFunc func(context.Context, string, server.Plugin) (server.Plugin, error)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(resolver ResolverFunc, plugins ...server.Plugin) (*provider, error) {
	self := new(provider)
	self.plugin = make(map[string]server.Plugin, len(plugins))
	self.task = make(map[string]*state, len(plugins))
	self.order = make([]string, 0, len(plugins))
	self.resolver = resolver
	self.Logger = logger.New(os.Stderr, logger.Term, false)

	// Add the plugins
	for _, plugin := range plugins {
		// Check plugin
		if plugin == nil {
			return nil, httpresponse.ErrInternalError.With("Plugin is nil")
		} else if !types.IsIdentifier(plugin.Name()) {
			return nil, httpresponse.ErrInternalError.Withf("Plugin name %q is not valid", plugin.Name())
		}

		// TODO: Don't use names, use labels!
		label := plugin.Name()
		if _, exists := self.plugin[label]; exists {
			return nil, httpresponse.ErrInternalError.Withf("Plugin %q already exists", plugin.Name())
		} else {
			self.plugin[label] = plugin
		}
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (provider *provider) MarshalJSON() ([]byte, error) {
	type jtask struct {
		Name        string        `json:"name"`
		Description string        `json:"description,omitempty"`
		Label       string        `json:"label,omitempty"`
		Plugin      server.Plugin `json:"plugin,omitempty"`
		Task        server.Task   `json:"task,omitempty"`
	}
	result := make([]jtask, 0, len(provider.task))
	for _, label := range provider.order {
		plugin := provider.plugin[label]
		result = append(result, jtask{
			Name:        plugin.Name(),
			Description: plugin.Description(),
			Label:       label,
			Plugin:      plugin,
			Task:        provider.task[label].Task,
		})
	}
	return json.Marshal(result)
}

func (provider *provider) String() string {
	data, err := json.MarshalIndent(provider, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a task from a label
func (provider *provider) Task(ctx context.Context, label string) server.Task {
	provider.Print(ctx, "Called Task for ", label)

	// If the task is already created, then return it
	if task, exists := provider.task[label]; exists {
		return task.Task
	}

	// If the plugin doesn't exist, return nil
	plugin, exists := provider.plugin[label]
	if !exists {
		return nil
	}

	// Resolve the plugin
	if provider.resolver != nil {
		var err error
		plugin, err = provider.resolver(withProvider(ctx, provider), label, plugin)
		if err != nil {
			provider.Print(ctx, "Error: ", label, ": ", err)
			return nil
		}
	}

	// Create the task
	task, err := plugin.New(withPath(ctx, label))
	if err != nil {
		provider.Print(ctx, "Error: ", label, ": ", err)
		return nil
	} else if task == nil {
		provider.Print(ctx, "Error: ", label, ": ", httpresponse.ErrInternalError.With("Task is nil"))
		return nil
	}

	// If it's a logger, replace the current logger
	if logger, ok := task.(server.Logger); ok && logger != nil {
		logger.Debugf(ctx, "Replacing logger with %q", label)
		provider.Logger = logger
	}

	// Set the task and order
	provider.task[label] = &state{Task: task}
	provider.order = append(provider.order, label)

	// Return the task
	return task
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Make all tasks
func (provider *provider) constructor(ctx context.Context) error {
	for label := range provider.plugin {
		if task := provider.Task(ctx, label); task == nil {
			return httpresponse.ErrConflict.Withf("Failed to create task %q", label)
		}
	}
	return nil
}
