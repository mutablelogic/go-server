package auth

import (
	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	TokenJar   TokenJar `hcl:"token_jar" description:"Persistent storage for tokens"`
	TokenBytes int      `hcl:"token_bytes" description:"Number of bytes in a token"`
	Bearer     bool     `hcl:"bearer" description:"Use bearer token for authorization"`
}

// Check interfaces are satisfied
var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName       = "auth-handler"
	defaultTokenBytes = 16
	defaultRootNme    = "root"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "token and group management for authentication and authorisation"
}

// Create a new task from the configuration
func (c Config) New() (server.Task, error) {
	return New(c)
}
