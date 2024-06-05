package provider

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"plugin"

	// Packages
	server "github.com/mutablelogic/go-server"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugin represents a plugin which can be loaded
type Plugin struct {
	Path string      `json:"path"`
	Meta *PluginMeta `json:"meta"`

	// Private fields
	plugin server.Plugin
}

// pluginProvider is a list of plugins
type pluginProvider struct {
	plugins map[string]*Plugin
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	configFunc = "Plugin"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(plugins ...server.Plugin) *pluginProvider {
	self := new(pluginProvider)
	self.plugins = make(map[string]*Plugin, len(plugins))

	// TODO

	return self
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LoadPluginsForPattern will load plugins from filesystem
// for a given glob pattern
func (p *pluginProvider) LoadPluginsForPattern(pattern string) error {
	plugins, err := loadPluginsForPattern(pattern)
	if err != nil {
		return err
	}

	// TODO

	// Return success
	return nil
}

// Create a configuration object for a plugin with a label
func (p *pluginProvider) New(name, label string) (server.Plugin, error) {
	// TODO
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (plugin *Plugin) String() string {
	data, _ := json.MarshalIndent(plugin, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// loadPluginsForPattern will load and return a list of plugins for a given glob pattern
func loadPluginsForPattern(pattern string) ([]*Plugin, error) {
	var result error
	var plugins []*Plugin

	// Seek plugins
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, ErrNotFound.Withf("No plugins found for pattern: %q", pattern)
	}

	// Load plugins, and create metadata object for the block
	for _, path := range files {
		plugin, err := pluginWithPath(path)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		meta, err := NewMeta(plugin)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		plugins = append(plugins, &Plugin{
			Path:   filepath.Clean(path),
			Meta:   meta,
			plugin: plugin,
		})
	}

	// Return any errors
	return plugins, result
}

// Create a new plugin from a filepath
func pluginWithPath(path string) (server.Plugin, error) {
	// Check path to make sure it's a regular file
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.Mode().IsRegular() {
		return nil, ErrBadParameter.Withf("Not a regular file: %q", path)
	}

	// Load the plugin
	if plugin, err := plugin.Open(path); err != nil {
		return nil, err
	} else if fn, err := plugin.Lookup(configFunc); err != nil {
		return nil, err
	} else if fn_, ok := fn.(func() server.Plugin); !ok {
		_ = fn.(func() server.Plugin)
		return nil, ErrInternalAppError.With("New returned nil: ", path)
	} else if config := fn_(); config == nil {
		return nil, ErrInternalAppError.With("New returned nil: ", path)
	} else {
		return config, nil
	}
}
