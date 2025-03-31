package pgqueue

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Pool      server.PG         `kong:"-"`           // Connection pool
	Router    server.HTTPRouter `kong:"-"`           // Which HTTP router to use
	Namespace *string           `help:"Namespace"`   // Namespace for queues and tickers
	Worker    *string           `help:"Worker name"` // The name of the worker
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	if c.Pool == nil {
		return nil, httpresponse.ErrBadRequest.With("missing connection pool")
	}

	// Add options
	opts := []pgqueue.Opt{}
	if c.Worker != nil {
		opts = append(opts, pgqueue.OptWorker(*c.Worker))
	}
	if c.Namespace != nil {
		opts = append(opts, pgqueue.OptNamespace(*c.Namespace))
	}

	// Create a new client
	client, err := pgqueue.New(ctx, c.Pool.Conn(), opts...)
	if err != nil {
		return nil, err
	}

	// TODO: Register HTTP handlers
	if c.Router != nil {
		// TODO: Register HTTP handlers
	}

	// Return the task
	return taskWith(client), nil
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return schema.SchemaName
}

func (c Config) Description() string {
	return "Postgresql task queue"
}
