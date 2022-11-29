package nginx

import (
	"context"
	"os"
	"path/filepath"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Path_   types.String            `json:"path,omitempty"`   // Path to the nginx binary
	Config_ types.String            `json:"config,omitempty"` // Path to the configuration file
	Prefix_ types.String            `json:"prefix,omitempty"` // Prefix for nginx configuration
	Env_    map[string]types.String `json:"env,omitempty"`    // Environment variable map
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

func (p Plugin) Flags() []string {
	result := make([]string, 0, 5)

	// Switch off daemon setting
	result = append(result, "-g", "daemon off;")

	// Check for configuration file
	if p.Config_ != "" {
		result = append(result, "-c", p.Config())
	}

	// Check for prefix path
	if p.Prefix_ != "" {
		result = append(result, "-p", string(p.Prefix_))
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
