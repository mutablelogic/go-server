package pgqueue

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	*pgqueue.Client
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func taskWith(queue *pgqueue.Client) *task {
	return &task{queue}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (*task) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
