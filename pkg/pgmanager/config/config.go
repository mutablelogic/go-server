package config

import (
	"context"

	// Packages
	pg "github.com/djthorpe/go-pg"
	server "github.com/mutablelogic/go-server"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager"
	pghandler "github.com/mutablelogic/go-server/pkg/pgmanager/handler"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Addr     string            `name:"host" env:"PG_HOST" default:"localhost" help:"Server address, <host> or <host>:<port>"`
	User     string            `env:"PG_USER" default:"${USER}" help:"Database user"`
	Pass     string            `env:"PG_PASSWORD" help:"User password"`
	Database string            `env:"PG_DATABASE" help:"Database name, uses username if not set"`
	SSLMode  string            `env:"PG_SSLMODE" enum:"default,disable,allow,prefer,require,verify-ca,verify-full" help:"SSL mode"`
	Trace    pg.TraceFn        `json:"-" kong:"-"`
	Router   server.HTTPRouter `json:"-" kong:"-"` // Which HTTP router to use
}

var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	opts := []pg.Opt{}
	if c.Addr != "" {
		opts = append(opts, pg.WithAddr(c.Addr))
	}
	if c.User != "" || c.Pass != "" {
		opts = append(opts, pg.WithCredentials(c.User, c.Pass))
	}
	if c.Database != "" {
		opts = append(opts, pg.WithDatabase(c.Database))
	}
	if c.SSLMode != "" && c.SSLMode != "default" {
		opts = append(opts, pg.WithSSLMode(c.SSLMode))
	}
	if c.Trace != nil {
		opts = append(opts, pg.WithTrace(c.Trace))
	}

	// Create  a connection pool
	pool, err := pg.NewPool(ctx, opts...)
	if err != nil {
		return nil, err
	} else if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		if manager, err := pgmanager.New(ctx, pool); err != nil {
			return nil, err
		} else {
			pghandler.RegisterRole(ctx, c.Router, schema.APIPrefix, manager)
			pghandler.RegisterDatabase(ctx, c.Router, schema.APIPrefix, manager)
			pghandler.RegisterSchema(ctx, c.Router, schema.APIPrefix, manager)
			pghandler.RegisterObject(ctx, c.Router, schema.APIPrefix, manager)
			pghandler.RegisterConnection(ctx, c.Router, schema.APIPrefix, manager)
			pghandler.RegisterTablespace(ctx, c.Router, schema.APIPrefix, manager)
		}
	}

	// Create a new task with the connection pool
	return NewTask(pool), nil
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return "pgpool"
}

func (c Config) Description() string {
	return "Postgresql Connection Pool"
}
