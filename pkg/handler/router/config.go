package router

import (
	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Services ServiceConfig `hcl:"services"`
}

type ServiceConfig map[string]struct {
	Service    server.ServiceEndpoints `hcl:"service"`
	Middleware []server.Middleware     `hcl:"middleware"`
}

// Ensure interfaces is implemented
var _ server.Plugin = Config{}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName      = "router"
	defaultCap       = 10
	pathSep          = "/"
	hostSep          = "."
	envRequestPrefix = "REQUEST_PREFIX"
	envServerName    = "SERVER_NAME"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "router for http requests"
}

// Create a new router from the configuration
func (c Config) New() (server.Task, error) {
	return New(c)
}
