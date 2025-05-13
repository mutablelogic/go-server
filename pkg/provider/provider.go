package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"slices"
	"strings"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	meta "github.com/mutablelogic/go-server/pkg/provider/meta"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type provider struct {
	// Plugin metadata
	protos map[string]*meta.Meta

	// Order of the plugins
	porder []string

	// Map labels to plugins
	plugin map[string]server.Plugin

	// Map labels to tasks
	task map[string]*state

	// Order that the tasks were created
	order []string

	// Map labels to resolvers
	resolvers map[string]server.PluginResolverFunc

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

func New(plugins ...server.Plugin) (*provider, error) {
	self := new(provider)
	self.plugin = make(map[string]server.Plugin, len(plugins))
	self.task = make(map[string]*state, len(plugins))
	self.order = make([]string, 0, len(plugins))
	self.resolvers = make(map[string]server.PluginResolverFunc, len(plugins))
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
		} else if label == providerLabel {
			return nil, httpresponse.ErrInternalError.Withf("Label %q is reserved", providerLabel)
		} else {
			self.porder = append(self.porder, label)
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

func (provider *provider) WriteConfig(w io.Writer) error {
	var buf bytes.Buffer
	for _, proto := range provider.protos {
		if err := proto.Write(&buf); err != nil {
			return err
		}
		buf.WriteRune('\n')
	}
	_, err := w.Write(buf.Bytes())
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all the plugins
func (provider *provider) Plugins() []server.Plugin {
	var result []server.Plugin
	for _, meta := range provider.protos {
		result = append(result, meta.New())
	}
	return result
}

// Return a task from a label
func (provider *provider) Task(ctx context.Context, label string) server.Task {
	provider.Debugf(ctx, "Called Task for %q", label)

	// If the task is already created, then return it
	if task, exists := provider.task[label]; exists {
		return task.Task
	}

	// If the plugin doesn't exist, return nil
	plugin, exists := provider.plugin[label]
	if !exists {
		provider.Print(ctx, label, ": ", httpresponse.ErrNotFound.Withf("Plugin %q not found", label))
		return nil
	}

	// Check for circular dependency
	if path := ref.Path(ctx); slices.Contains(path, label) {
		provider.Print(ctx, httpresponse.ErrInternalError.Withf("circular dependency for %s -> %s", strings.Join(path, " -> "), label))
		return nil
	}

	// modify ctx to append path
	ctx = ref.WithPath(ctx, label)

	// Resolve the plugin
	if fn := provider.resolvers[label]; fn != nil {
		if err := fn(ctx, label, plugin); err != nil {
			provider.Print(ctx, label, ": ", err)
			return nil
		}
	}

	// Create the task
	provider.Debugf(ctx, "Creating a new task %q", label)
	task, err := plugin.New(ctx)
	if err != nil {
		provider.Print(ctx, label, ": ", err)
		return nil
	} else if task == nil {
		provider.Print(ctx, label, ": ", httpresponse.ErrInternalError.With("Task is nil"))
		return nil
	}

	// If it's a logger, replace the current logger
	if logger, ok := task.(server.Logger); ok && logger != nil {
		provider.Logger = logger
	}

	// Set the task and order
	provider.task[label] = &state{Task: task}
	provider.order = append(provider.order, label)

	// Return the task
	return task
}
