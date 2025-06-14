package config

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/auth"
	handler "github.com/mutablelogic/go-server/pkg/auth/handler"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Pool   server.PG         `kong:"-"` // Connection pool
	Router server.HTTPRouter `kong:"-"` // Which HTTP router to use
}

var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	// Check for connection pool
	if c.Pool == nil {
		return nil, httpresponse.ErrInternalError.With("missing connection pool")
	}

	// Create  an auth manager
	manager, err := auth.New(ctx, c.Pool.Conn(), []auth.Opt{}...)
	if err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		handler.RegisterUser(ctx, c.Router, schema.APIPrefix, manager)
		handler.RegisterToken(ctx, c.Router, schema.APIPrefix, manager)
		handler.RegisterAuth(ctx, c.Router, schema.APIPrefix, manager)
	}

	// Create a new task with the connection pool
	return NewTask(c.Pool.Conn()), nil
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return schema.SchemaName
}

func (c Config) Description() string {
	return "Authorization manager"
}
