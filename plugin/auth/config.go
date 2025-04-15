package pg

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/auth"
	handler "github.com/mutablelogic/go-server/pkg/auth/handler"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Pool   server.PG         `kong:"-"`                                    // Connection pool
	Router server.HTTPRouter `kong:"-"`                                    // Which HTTP router to use
	Prefix string            `default:"${AUTH_PREFIX}" help:"Path prefix"` // HTTP Path Prefix
}

var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	// Create  an auth manager
	manager, err := auth.New(ctx, c.Pool.Conn(), []auth.Opt{}...)
	if err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		handler.RegisterUser(ctx, c.Router, c.Prefix, manager)
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
