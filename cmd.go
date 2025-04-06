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
	// Return the context
	Context() context.Context

	// Return the debug mode
	GetDebug() bool

	// Return the endpoint
	GetEndpoint(paths ...string) *url.URL

	// Return the HTTP client options
	GetClientOpts() []client.ClientOpt
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CmdOffsetLimit struct {
	Offset uint64  `name:"offset" help:"List item offset"`
	Limit  *uint64 `name:"limit" help:"List item limit"`
}
