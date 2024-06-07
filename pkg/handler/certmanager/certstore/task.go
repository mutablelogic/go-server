package certstore

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.Task = (*certstore)(nil)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the label
func (task *certstore) Label() string {
	// TODO
	return defaultName
}

// Run the task until the context is cancelled
func (task *certstore) Run(ctx context.Context) error {
	var result error

	// Run the task until cancelled
	<-ctx.Done()

	// Return any errors
	return result
}
