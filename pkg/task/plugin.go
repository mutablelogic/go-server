package task

import (
	"context"
	"fmt"
	"path/filepath"
	"plugin"

	// Package imports
	"github.com/hashicorp/go-multierror"
	iface "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugins is a map of all registered plugins
type Plugins map[string]iface.Plugin

// Plugin creates tasks from a configuration
type Plugin struct {
	Name_  types.String `json:"name,omitempty"`
	Label_ types.String `json:"label,omitempty"`
}

// Compile time check
var _ iface.Plugin = (*Plugin)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	funcConfig = "Config"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register adds a plugin to the map of plugins. It will return errors if the
// name or label is invalid, or the plugin with the same name already exists.
func (p Plugins) Register(v ...iface.Plugin) error {
	var result error
	if len(v) == 0 {
		return ErrBadParameter.With("Register")
	}
	for _, plugin := range v {
		if name := plugin.Name(); name == "" {
			result = multierror.Append(result, ErrBadParameter.Withf("Plugin without name"))
		} else if !types.IsIdentifier(name) {
			result = multierror.Append(result, ErrBadParameter.Withf("Plugin with invalid name: %q", name))
		} else if _, exists := p[name]; exists {
			return multierror.Append(result, ErrDuplicateEntry.Withf("Plugin with duplicate name: %q", name))
		} else {
			p[name] = plugin
		}
	}

	// Return any errors
	return result
}

// LoadPluginsForPattern will load and return a map of plugins for a given glob pattern,
// keyed against the plugin name.
func (p Plugins) LoadPluginsForPattern(pattern string) error {
	var result error

	// Seek plugins
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Load plugins
	for _, path := range files {
		plugin, err := PluginWithPath(path)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		// Check for duplicate plugins
		name := plugin.Name()
		if _, exists := p[name]; exists {
			result = multierror.Append(result, ErrDuplicateEntry.Withf("Duplicate plugin: %q", name))
			continue
		}

		// Set plugin
		p[name] = plugin
	}

	// Return any errors
	return result
}

// Create a new plugin from a filepath
func PluginWithPath(path string) (iface.Plugin, error) {
	if plugin, err := plugin.Open(path); err != nil {
		return nil, err
	} else if fn, err := plugin.Lookup(funcConfig); err != nil {
		return nil, err
	} else if fn_, ok := fn.(func() iface.Plugin); !ok {
		return nil, ErrInternalAppError.With("New returned nil: ", path)
	} else if config := fn_(); config == nil {
		return nil, ErrInternalAppError.With("New returned nil: ", path)
	} else {
		return config, nil
	}
}

// Return error if the plugin does not have a name or label
func (p Plugin) HasNameLabel() error {
	if p.Name() == "" {
		return ErrBadParameter.With("Plugin has no name")
	}
	if !types.IsIdentifier(p.Name()) {
		return ErrBadParameter.Withf("Invalid plugin name: %q", p.Name())
	}
	if p.Label() == "" {
		return ErrBadParameter.Withf("Plugin %q has no label", p.Name())
	}
	if !types.IsIdentifier(p.Label()) {
		return ErrBadParameter.Withf("Invalid plugin label: %q", p.Label())
	}
	return nil
}

// Create a new default task from the plugin
func (p Plugin) New(c context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}
	return NewTask(ctx.WithNameLabel(c, p.Name(), p.Label()), provider)
}

// Create a new default task from the plugin
func (p Plugin) WithLabel(label string) iface.Plugin {
	p.Label_ = types.String(label)
	fmt.Println("p", p)
	return p
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return name of the plugin
func (p Plugin) Name() string {
	return string(p.Name_)
}

// Return label of the plugin
func (p Plugin) Label() string {
	return string(p.Label_)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return string representation for a plugin
func (p Plugin) String() string {
	str := "<plugin"
	if v := p.Name(); v != "" {
		str += fmt.Sprintf(" name=%q", v)
	}
	if v := p.Label(); v != "" {
		str += fmt.Sprintf(" label=%q", v)
	}
	return str + ">"
}
