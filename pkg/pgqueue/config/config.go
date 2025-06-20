package config

import (
	"context"
	"runtime"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	handler "github.com/mutablelogic/go-server/pkg/pgqueue/handler"
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

var _ server.Plugin = Config{}

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

	// Create a new queue manager
	manager, err := pgqueue.NewManager(ctx, c.Pool.Conn(), opts...)
	if err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		handler.Register(ctx, c.Router, schema.APIPrefix, manager)
	}

	// Return the task
	return NewTask(manager, uint(runtime.NumCPU()))
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return schema.SchemaName
}

func (c Config) Description() string {
	return "PostgreSQL Task Queue"
}
