package server

import (
	"context"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Cmd represents the command line interface context
type Cmd interface {
	// Run the command
	Run() error

	// Return the context
	Context() context.Context

	// Return the debug mode
	GetDebug() DebugLevel

	// Return the endpoint
	GetEndpoint(paths ...string) *url.URL

	// Return the HTTP client options
	GetClientOpts() []client.ClientOpt
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

type DebugLevel uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	None DebugLevel = iota
	Debug
	Trace
)
