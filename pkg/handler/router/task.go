package router

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Ensure interfaces is implemented
var _ server.Task = (*router)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the label for the task
func (router *router) Label() string {
	// TODO
	return defaultName
}

// Run the router until the context is cancelled
func (router *router) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
