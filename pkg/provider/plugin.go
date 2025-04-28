package provider

import (
	"errors"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	meta "github.com/mutablelogic/go-server/pkg/provider/meta"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	configFunc = "Plugin"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewWithPlugins(resolver ResolverFunc, paths ...string) (server.Provider, error) {
	self := new(provider)

	// Load plugins
	plugins, err := loadPluginsForPattern(paths...)
	if err != nil {
		return nil, err
	}

	// Create the prototype map
	self.protos = make(map[string]*meta.Meta, len(plugins))
	self.plugin = make(map[string]server.Plugin, len(plugins))
	self.task = make(map[string]*state, len(plugins))
	self.resolver = resolver
	self.Logger = logger.New(os.Stderr, logger.Term, false)
	for _, plugin := range plugins {
		name := plugin.Name()
		if _, exists := self.protos[name]; exists {
			return nil, httpresponse.ErrBadRequest.Withf("Duplicate plugin: %q", name)
		} else if !types.IsIdentifier(name) {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid plugin: %q", name)
		} else if meta, err := meta.New(plugin, name); err != nil {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid plugin: %q", name)
		} else {
			self.protos[name] = meta
		}
	}

	// Return success
	return self, nil
}

// Create a "concrete" plugin from a prototype and a label, using the function to "hook"
// any values into the plugin
func (provider *provider) Load(name, label string, fn func(server.Plugin)) error {
	proto, exists := provider.protos[name]
	if !exists {
		return httpresponse.ErrBadRequest.Withf("Plugin not found: %q", name)
	}

	// Make the label
	if label != "" {
		if !types.IsIdentifier(label) {
			return httpresponse.ErrBadRequest.Withf("Invalid label: %q", label)
		} else {
			label = name
		}
	} else {
		label = strings.Join([]string{label, name}, ".")
	}

	// Register the plugin with the label
	if _, exists := provider.plugin[label]; exists {
		return httpresponse.ErrBadRequest.Withf("Plugin already exists: %q", label)
	} else {
		// Create a plugin from the prototype
		plugin := proto.New()
		if fn != nil {
			fn(plugin)
		}
		provider.plugin[label] = plugin
		provider.porder = append(provider.porder, label)
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// loadPluginsForPattern will load plugins from filesystem
// for a given glob pattern
func loadPluginsForPattern(pattern ...string) ([]server.Plugin, error) {
	var result []server.Plugin
	var errs error

	// Seek plugins
	for _, p := range pattern {
		files, err := filepath.Glob(p)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			return nil, httpresponse.ErrBadRequest.Withf("No plugins found for pattern: %q", pattern)
		}

		// Load plugins
		for _, path := range files {
			plugin, err := pluginWithPath(path)
			if err != nil {
				errs = errors.Join(errs, err)
			} else {
				result = append(result, plugin)
			}
		}
	}

	// Return any errors
	return result, errs
}

// Create a new plugin from a filepath
func pluginWithPath(path string) (server.Plugin, error) {
	// Check path to make sure it's a regular file
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.Mode().IsRegular() {
		return nil, httpresponse.ErrBadRequest.Withf("Not a regular file: %q", path)
	}

	// Load the plugin
	if plugin, err := plugin.Open(path); err != nil {
		return nil, err
	} else if fn, err := plugin.Lookup(configFunc); err != nil {
		return nil, err
	} else if fn_, ok := fn.(func() server.Plugin); !ok {
		_ = fn.(func() server.Plugin)
		return nil, httpresponse.ErrInternalError.With("New returned nil: ", path)
	} else if config := fn_(); config == nil {
		return nil, httpresponse.ErrInternalError.With("New returned nil: ", path)
	} else {
		return config, nil
	}
}
