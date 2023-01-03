package main

import (
	"context"

	// Packages
	iface "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/npm/router"
	static "github.com/mutablelogic/go-server/pkg/httpserver/static"
	task "github.com/mutablelogic/go-server/pkg/task"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Plugin for serving a file system as a static web server
type plugin struct {
	task.Plugin
}

var _ iface.Plugin = (*plugin)(nil)

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "router-frontend"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p plugin) New(parent context.Context, provider iface.Provider) (iface.Task, error) {
	// Check parameters
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Return static file server with the embedded files
	return static.NewWithPlugin(p, router.Dist, "dist")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Config() iface.Plugin {
	return plugin{}
}

func (p plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}
