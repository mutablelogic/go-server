package provider

import (
	"context"
	"errors"
	"fmt"

	// Packages
	hcl "github.com/mutablelogic/go-server/pkg/hcl"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Provider represents all the resources for a given configuration
type provider struct {
	plugin map[string]*Plugin
}

// Provider is a resource itself, also provides logger
type Provider interface {
	hcl.Resource
	Logger
}

// Ensure that provider implements the Provider interface
var _ Provider = (*provider)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 20
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new tasks object and load plugins
func NewPluginsForPattern(ctx context.Context, patterns ...string) (*provider, error) {
	var result error

	// Create tasks
	tasks := &provider{
		plugin: make(map[string]*Plugin, defaultCapacity),
	}

	for _, pattern := range patterns {
		plugins, err := LoadPluginsForPattern(pattern)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		for _, plugin := range plugins {
			if _, exists := tasks.plugin[plugin.Name]; exists {
				result = errors.Join(result, ErrDuplicateEntry.Withf("%q", plugin.Name))
			} else {
				tasks.plugin[plugin.Name] = plugin
			}
		}
	}

	// Return success
	return tasks, result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a new resource from a block, and optionally set the resource label
func (self *provider) NewResource(name string, label hcl.Label) (hcl.Resource, error) {
	return nil, ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (self *provider) String() string {
	str := "<provider"
	for _, plugin := range self.plugin {
		str += "\n  " + fmt.Sprint(plugin)
	}
	return str + ">"
}
