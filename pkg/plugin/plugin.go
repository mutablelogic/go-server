package plugin

import (
	"debug/elf"
	"debug/macho"
	"errors"
	"os"
	"path/filepath"
	"plugin"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Plugin is the interface that plugins must implement to be loaded by the server.
type Plugin interface {
	schema.Provider
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	configFunc = "Provider"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// PluginsForPattern will load plugins from filesystem for a given glob pattern
func PluginsForPattern(pattern ...string) ([]Plugin, error) {
	var result []Plugin
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

// isGoPlugin reports whether the file at path is a shared library / dynamic
// object that can be loaded as a Go plugin, using the stdlib debug/elf and
// debug/macho packages to inspect the binary header.
func isGoPlugin(path string) bool {
	// Mach-O (macOS): only MH_DYLIB and MH_BUNDLE are valid plugin targets.
	if f, err := macho.Open(path); err == nil {
		defer f.Close()
		return f.Type == macho.TypeDylib || f.Type == macho.TypeBundle
	}
	// ELF (Linux/*BSD): only ET_DYN shared objects.
	if f, err := elf.Open(path); err == nil {
		defer f.Close()
		return f.Type == elf.ET_DYN
	}
	return false
}

// Create a new plugin from a filepath
func pluginWithPath(path string) (Plugin, error) {
	// Check path to make sure it's a regular file
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.Mode().IsRegular() {
		return nil, httpresponse.ErrBadRequest.Withf("Not a regular file: %q", path)
	}

	// Inspect the binary header before calling plugin.Open — loading a
	// non-plugin binary (e.g. a regular executable) triggers a fatal runtime
	// error that cannot be caught by recover.
	if !isGoPlugin(path) {
		return nil, httpresponse.ErrBadRequest.Withf("Not a Go plugin (wrong binary type): %q", path)
	}

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	fn, err := p.Lookup(configFunc)
	if err != nil {
		return nil, err
	}
	fn_, ok := fn.(func() Plugin)
	if !ok {
		return nil, httpresponse.ErrInternalError.Withf("Provider has wrong signature in %q", path)
	}
	config := fn_()
	if config == nil {
		return nil, httpresponse.ErrInternalError.Withf("Provider returned nil in %q", path)
	}
	return config, nil
}
