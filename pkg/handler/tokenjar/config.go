package tokenjar

import (
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	DataPath      string        `hcl:"datapath" description:"Path to persistent data"`
	WriteInterval time.Duration `hcl:"write-interval" description:"Interval to write data to disk"`
}

// Check interfaces are satisfied
var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "tokenjar-handler"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "on-disk token persistence"
}

// Create a new task from the configuration
func (c Config) New() (server.Task, error) {
	return New(c)
}
