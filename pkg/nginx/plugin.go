package nginx

import (
	"context"

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
	Path_   types.Path              `json:"path,omitempty"`   // Path to the nginx binary
	Config_ types.Path              `json:"config,omitempty"` // Path to the configuration file
	Prefix_ types.Path              `json:"prefix,omitempty"` // Prefix for nginx configuration
	Env_    map[string]types.String `json:"env,omitempty"`    // Environment variable map

	path  string
	flags []string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "nginx"
	defaultPath = "/usr/sbin/nginx"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Switch off daemon setting
	p.flags = append(p.flags, "-g", "daemon off;")

	// Check for executable path
	if p.Path_ == "" {
		p.Path_ = types.Path(defaultPath)
	}
	if abs, err := p.Path_.AbsFile(); err != nil {
		return nil, err
	} else {
		p.path = abs
	}

	// Check for configuration file
	if p.Config_ != "" {
		if abs, err := p.Config_.AbsFile(); err != nil {
			return nil, err
		} else {
			p.flags = append(p.flags, "-c", abs)
		}
	}

	// Check for prefix path
	if p.Prefix_ != "" {
		if abs, err := p.Config_.AbsDir(); err != nil {
			return nil, err
		} else {
			p.flags = append(p.flags, "-p", abs)
		}
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
