package config

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (*task) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
