package nginx

import (
	"context"
	"os"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Path_   types.String `json:"path,omitempty"`   // Path to the nginx binary
	Config_ types.String `json:"config,omitempty"` // Path to the configuration file
	Prefix_ types.String `json:"prefix,omitempty"` // Prefix for nginx configuration

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
		p.Path_ = types.String(defaultPath)
	}
	if stat, err := os.Stat(string(p.Path_)); err != nil {
		return nil, err
	} else if !IsExecAny(stat.Mode()) {
		return nil, ErrBadParameter.Withf("path is not executable: %q", p.Path_)
	} else {
		p.path = string(p.Path_)
	}

	// Check for configuration file
	if p.Config_ != "" {
		if stat, err := os.Stat(string(p.Config_)); err != nil {
			return nil, err
		} else if stat.Mode().IsRegular() == false {
			return nil, ErrBadParameter.Withf("config is not a regular file: %q", p.Config_)
		} else {
			p.flags = append(p.flags, "-c", string(p.Config_))
		}
	}

	// Check for prefix path
	if p.Prefix_ != "" {
		if stat, err := os.Stat(string(p.Prefix_)); err != nil {
			return nil, err
		} else if stat.Mode().IsDir() == false {
			return nil, ErrBadParameter.Withf("prefix is not a directory: %q", p.Prefix_)
		} else {
			p.flags = append(p.flags, "-p", string(p.Prefix_))
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
