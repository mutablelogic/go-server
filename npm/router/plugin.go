package routerFrontend

import (
	"embed"

	// Packages
	server "github.com/mutablelogic/go-server"
	static "github.com/mutablelogic/go-server/pkg/handler/static"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// No configuration
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

//go:embed build
var dist embed.FS

const (
	defaultName = "router-frontend"
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Returns the plugin
func Plugin() server.Plugin {
	return Config{}
}

// Return the unique name for the plugin
func (c Config) Name() string {
	return defaultName
}

// Return a description of the plugin
func (c Config) Description() string {
	return "frontend for the router handler"
}

// Create a task from a plugin
func (c Config) New() (server.Task, error) {
	return static.Config{
		FS:         dist,
		DirPrefix:  "build",
		DirListing: true,
	}.New()
}
