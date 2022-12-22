package task

import (
	"errors"
	"fmt"
	"reflect"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	iface "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type graph struct {
	edges   map[string][]string
	plugins map[string]iface.Plugin
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	typeTask = reflect.TypeOf(types.Task{})
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewGraph creates a graph of plugins, so that dependencies can be resolved
// in the correct order and there are no circular references
func NewGraph(plugins ...iface.Plugin) *graph {
	graph := new(graph)
	graph.plugins = make(map[string]iface.Plugin, len(plugins))
	graph.edges = make(map[string][]string, len(plugins))

	// Add dependencies for each plugin
	for _, plugin := range plugins {
		graph.addEdges(plugin)
	}

	// Return the graph
	return graph
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (graph *graph) String() string {
	str := "<provider.graph"
	for key, refs := range graph.edges {
		str += fmt.Sprintf(" %q => %q", key, refs)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Resolve returns a list of plugins in the correct order, so that dependencies
// are satisfied correctly. If there are any circular dependencies, an error is
// returned.
func (graph *graph) Resolve() ([]iface.Plugin, error) {
	var result error

	// Create a list of plugins in the correct order
	resolved := make(map[string]bool, len(graph.edges))
	unresolved := make(map[string]bool, len(graph.edges))
	order := make([]string, 0, len(graph.edges))
	for key := range graph.edges {
		var err error
		if _, exists := resolved[key]; exists {
			continue
		} else if order, err = graph.resolve(key, order, resolved, unresolved); err != nil {
			result = multierror.Append(result, errors.New(key+": "+err.Error()))
		}
	}

	// Return any errors (circular references)
	if result != nil {
		return nil, result
	}

	// Create a list of plugins in the correct order
	plugins := make([]iface.Plugin, 0, len(order))
	for _, key := range order {
		if plugin, exists := graph.plugins[key]; !exists {
			result = multierror.Append(result, ErrNotFound.Withf("unresolved reference: %q", key))
		} else {
			plugins = append(plugins, plugin)
		}
	}

	// Return any errors (unresolved references)
	return plugins, result
}

// KeyForPlugin returns the canonical key for a plugin
func KeyForPlugin(plugin iface.Plugin) string {
	return plugin.Name() + "." + plugin.Label()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// addEdges creates a list of dependencies for each plugin
func (graph *graph) addEdges(plugin iface.Plugin) {
	key := KeyForPlugin(plugin)
	graph.plugins[key] = plugin
	graph.edges[key] = []string{}
	resolveRef(key, reflect.ValueOf(plugin), func(ref string, _ reflect.Value) error {
		// If ref is not empty, then add a dependency
		if ref != "" {
			graph.edges[key] = append(graph.edges[key], ref)
		}
		return nil
	})
}

// resolve recursively resolves the dependencies for a plugin
func (graph *graph) resolve(key string, order []string, resolved, unresolved map[string]bool) ([]string, error) {
	unresolved[key] = true
	for _, ref := range graph.edges[key] {
		if _, exists := resolved[ref]; !exists {
			if _, exists := unresolved[ref]; exists {
				return order, ErrOutOfOrder.Withf("Circular dependency: %q -> %q", key, ref)
			}
			var err error
			if order, err = graph.resolve(ref, order, resolved, unresolved); err != nil {
				return order, err
			}
		}
	}
	resolved[key] = true
	order = append(order, key)
	delete(unresolved, key)
	return order, nil
}

// resolveRef resolves the references in the plugin, to build a dependency graph
func resolveRef(key string, v reflect.Value, fn func(string, reflect.Value) error) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		if v.Type() == typeTask {
			return fn(v.Interface().(types.Task).Ref, v)
		}
		for i := 0; i < v.NumField(); i++ {
			if err := resolveRef(key, v.Field(i), fn); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, mapkey := range v.MapKeys() {
			if err := resolveRef(key, v.MapIndex(mapkey), fn); err != nil {
				return err
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := resolveRef(key, v.Index(i), fn); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}
