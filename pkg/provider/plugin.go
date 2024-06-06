package provider

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugin represents a plugin which can be loaded
type pluginMeta struct {
	Path string      `json:"path,omitempty"`
	Name string      `json:"name"`
	Meta *PluginMeta `json:"meta"`

	// Private fields
	plugin server.Plugin
}

// pluginProvider is a list of plugins and configurations
type pluginProvider struct {
	plugins map[string]*pluginMeta
	labels  map[types.Label]server.Plugin
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	configFunc = "Plugin"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(plugins ...server.Plugin) (*pluginProvider, error) {
	self := new(pluginProvider)
	self.plugins = make(map[string]*pluginMeta, len(plugins))
	self.labels = make(map[types.Label]server.Plugin, len(plugins))

	var result error
	for _, plugin := range plugins {
		if plugin_, err := NewPlugin(plugin, ""); err != nil {
			result = errors.Join(result, err)
		} else if _, exists := self.plugins[plugin_.Name]; exists {
			result = errors.Join(result, ErrDuplicateEntry.With(plugin_.Name))
		} else {
			self.plugins[plugin_.Name] = plugin_
		}
	}

	// Return any errors
	return self, result
}

func NewPlugin(v server.Plugin, path string) (*pluginMeta, error) {
	meta, err := NewPluginMeta(v)
	if err != nil {
		return nil, err
	}
	return &pluginMeta{
		Path:   path,
		Name:   v.Name(),
		Meta:   meta,
		plugin: v,
	}, nil
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

	var result error
	for _, plugin := range plugins {
		if plugin_, err := NewPlugin(plugin, ""); err != nil {
			result = errors.Join(result, err)
		} else if _, exists := p.plugins[plugin_.Name]; exists {
			result = errors.Join(result, ErrDuplicateEntry.With(plugin_.Name))
		} else {
			p.plugins[plugin_.Name] = plugin_
		}
	}

	// Return success
	return nil
}

// Create a configuration object for a plugin with label parts
func (p *pluginProvider) New(name string, suffix ...string) (server.Plugin, error) {
	// Get the plugin
	plugin, exists := p.plugins[name]
	if !exists {
		return nil, ErrNotFound.Withf("plugin %q", name)
	}

	// Create the label
	label := types.NewLabel(name, suffix...)
	if label == "" {
		return nil, ErrBadParameter.Withf("invalid label with suffix %q", strings.Join(suffix, types.LabelSeparator))
	}

	// Check for existing label
	if _, exists := p.labels[label]; exists {
		return nil, ErrDuplicateEntry.With(label)
	} else {
		p.labels[label] = plugin.new()
	}

	// Create a new configuration
	return p.labels[label], nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (plugin *pluginMeta) String() string {
	data, _ := json.MarshalIndent(plugin, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// new makes a new copy of the plugin
func (plugin *pluginMeta) new() server.Plugin {
	rt := reflect.TypeOf(plugin.plugin)
	return reflect.New(rt).Interface().(server.Plugin)
}

// loadPluginsForPattern will load and return a list of plugins for a given glob pattern
func loadPluginsForPattern(pattern string) ([]server.Plugin, error) {
	var plugins []server.Plugin

	// Seek plugins
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, ErrNotFound.Withf("No plugins found for pattern: %q", pattern)
	}

	// Load plugins, and create metadata object for the block
	var result error
	for _, path := range files {
		plugin, err := pluginWithPath(path)
		if err != nil {
			result = errors.Join(result, err)
		} else {
			plugins = append(plugins, plugin)
		}
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
