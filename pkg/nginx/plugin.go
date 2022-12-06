package nginx

import (
	"context"
	"os"
	"path/filepath"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Path_      types.String            `json:"path,omitempty"`      // Path to the nginx binary
	Config_    types.String            `json:"config,omitempty"`    // Path to the configuration file
	Prefix_    types.String            `json:"prefix,omitempty"`    // Prefix for nginx configuration
	Available_ types.String            `json:"available,omitempty"` // Path to available configurations
	Enabled_   types.String            `json:"enabled,omitempty"`   // Path to enabled configurations
	Env_       map[string]types.String `json:"env,omitempty"`       // Environment variable map
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "nginx"
	defaultPath = defaultName
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Return task from plugin
	return NewWithPlugin(p)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

func (p Plugin) Path() string {
	if p.Path_ == "" {
		return defaultPath
	} else {
		return string(p.Path_)
	}
}

func (p Plugin) Config() string {
	if p.Config_ == "" {
		return ""
	} else if !filepath.IsAbs(string(p.Config_)) {
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, string(p.Config_))
		}
	}
	return string(p.Config_)
}

func (p Plugin) Prefix() string {
	if p.Prefix_ == "" {
		return ""
	} else if !filepath.IsAbs(string(p.Prefix_)) {
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, string(p.Prefix_))
		}
	}
	return string(p.Prefix_)
}

func (p Plugin) Available() string {
	if p.Available_ == "" {
		return ""
	} else if !filepath.IsAbs(string(p.Available_)) {
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, string(p.Available_))
		}
	}
	return string(p.Available_)
}

func (p Plugin) Enabled() string {
	if p.Enabled_ == "" {
		return ""
	} else if !filepath.IsAbs(string(p.Enabled_)) {
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, string(p.Enabled_))
		}
	}
	return string(p.Enabled_)
}

func (p Plugin) Flags() []string {
	result := make([]string, 0, 5)

	// Switch off daemon setting
	result = append(result, "-g", "daemon off;")

	// Check for configuration file
	if config := p.Config(); config != "" {
		result = append(result, "-c", config)
	}

	// Check for prefix path
	if prefix := p.Prefix(); prefix != "" {
		result = append(result, "-p", prefix)
	}

	// Return flags
	return result
}

func (p Plugin) Env() map[string]string {
	result := make(map[string]string, len(p.Env_))
	for k, v := range p.Env_ {
		result[k] = string(v)
	}
	return result
}
