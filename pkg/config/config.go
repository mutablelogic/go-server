package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plugin"

	// Modules
	. "github.com/djthorpe/go-server"
	multierror "github.com/hashicorp/go-multierror"
	yaml "gopkg.in/yaml.v3"
	//server "github.com/djthorpe/go-server/pkg/server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	//Server    *server.Config    `yaml:"server"`
	Plugins  []string          `yaml:"plugins"`
	Handlers map[string]string `yaml:"handlers"`
	Config   map[string]interface{}

	plugins map[string]*handler
}

type handler struct {
	Path   string
	Prefix string
	Plugin Plugin
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	funcName = "Name"
	funcNew  = "New"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(path string) (*Config, error) {
	this := new(Config)

	// Read configuration file
	if r, err := os.Open(path); err != nil {
		return nil, err
	} else {
		defer r.Close()
		dec := yaml.NewDecoder(r)
		if err := dec.Decode(&this); err != nil {
			return nil, err
		}
	}

	// When there are no plugins, fail with error
	if len(this.Plugins) == 0 {
		return nil, ErrBadParameter.With("plugins")
	} else {
		this.plugins = make(map[string]*handler, len(this.Plugins))
	}

	// Read plugin names
	var result error
	for _, path := range this.Plugins {
		if path := PluginPath(path); path == "" {
			result = multierror.Append(result, ErrNotFound.With("Plugin: ", path))
		} else if stat, err := os.Stat(path); err != nil {
			result = multierror.Append(result, ErrNotFound.With("Plugin: ", path))
		} else if stat.Mode().IsRegular() == false {
			result = multierror.Append(result, ErrBadParameter.With("Plugin: ", path))
		} else if name, err := GetPluginName(path); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Plugin: ", err))
		} else {
			this.plugins[name] = &handler{path, this.Handlers[name], nil}
		}
	}

	// Warn when a handler is not associated with a plugin name
	for handler := range this.Handlers {
		if this.plugins[handler] == nil {
			result = multierror.Append(result, ErrNotFound.With("Plugin: ", handler))
		}
	}

	// Return any errors
	if result != nil {
		return nil, result
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	name := filepath.Base(flag.CommandLine.Name())
	fmt.Fprintf(w, "%s: Monolith server\n", name)
	fmt.Fprintf(w, "\nUsage:\n")
	fmt.Fprintf(w, "  %s <flags> config.yaml\n", name)

	fmt.Fprintln(w, "\nFlags:")
	flag.PrintDefaults()

	fmt.Fprintln(w, "\nVersion:")
	PrintVersion(w)
	fmt.Fprintln(w, "")
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// GetPluginName returns the module name from a file path
func GetPluginName(path string) (string, error) {
	plugin, err := plugin.Open(path)
	if err != nil {
		return "", err
	}
	// Return module name
	if fn, err := plugin.Lookup(funcName); err != nil {
		return "", err
	} else if name := fn.(func() string)(); name == "" {
		return "", ErrInternalAppError.With("Name returned nil: ", path)
	} else {
		return name, nil
	}
}

// PluginPath returns absolute path to a plugin or empty string
// if it can't be located. Defaults to current working directory.
func PluginPath(path string) string {
	if abs, err := filepath.Abs(path); err != nil {
		return ""
	} else {
		return abs
	}
}
