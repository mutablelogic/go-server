package provider

import (
	"errors"
	"os"
	"path/filepath"
	"plugin"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	configFunc = "Plugin"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LoadPluginsForPattern will load plugins from filesystem
// for a given glob pattern
func LoadPluginsForPattern(pattern ...string) ([]server.Plugin, error) {
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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

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
