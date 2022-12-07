package gateway

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
	Nginx_  types.Task   `json:"nginx,omitempty"`  // Path to the nginx plugin
	Prefix_ types.String `json:"prefix,omitempty"` // Prefix for API, defaults to /nginx/v1
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName   = "nginx-gateway"
	defaultPrefix = "/nginx/v1"
	readScope     = "nginx.v1/read"
	writeScope    = "nginx.v1/write"
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

func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(defaultName, label),
	}
}

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

func (p Plugin) Task() iface.Task {
	return p.Nginx_.Task
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
