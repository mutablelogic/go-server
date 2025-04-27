package config

import (
	"context"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	logger "github.com/mutablelogic/go-server/pkg/logger"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	server.Logger
}

var _ server.Task = (*task)(nil)
var _ server.Logger = (*task)(nil)
var _ server.HTTPMiddleware = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newTaskWith(c Config) *task {
	return &task{
		Logger: logger.New(os.Stderr, logger.Term, c.Debug),
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *task) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
