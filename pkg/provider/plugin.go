package provider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	// Packages
	hcl "github.com/mutablelogic/go-server/pkg/hcl"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugin represents a plugin which can be loaded
type Plugin struct {
	Path string
	Name string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	funcBlock = "Block"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LoadPluginsForPattern will load and return a list of plugins for a given glob pattern
func LoadPluginsForPattern(pattern string) ([]*Plugin, error) {
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
		plugins = append(plugins, &Plugin{
			Path: filepath.Clean(path),
			Name: plugin.Name(),
		})
	}

	// Return any errors
	return plugins, result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (self *Plugin) String() string {
	str := "<plugin"
	if self.Name != "" {
		str += fmt.Sprintf(" name=%q", self.Name)
	}
	if self.Path != "" {
		str += fmt.Sprintf(" path=%q", self.Path)
	}
	/*	if self.Meta != nil && self.Meta.Description != "" {
		str += fmt.Sprintf(" description=%q", self.Meta.Description)
	}*/
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Create a new plugin from a filepath
func pluginWithPath(path string) (hcl.Block, error) {
	// Check path to make sure it's a regular file
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.Mode().IsRegular() {
		return nil, ErrBadParameter.Withf("Not a regular file: %q", path)
	}
	if plugin, err := plugin.Open(path); err != nil {
		return nil, err
	} else if fn, err := plugin.Lookup(funcBlock); err != nil {
		return nil, err
	} else if fn_, ok := fn.(func() hcl.Block); !ok {
		_ = fn.(func() hcl.Block)
		return nil, ErrInternalAppError.With("New returned nil: ", path)
	} else if config := fn_(); config == nil {
		return nil, ErrInternalAppError.With("New returned nil: ", path)
	} else {
		return config, nil
	}
}
