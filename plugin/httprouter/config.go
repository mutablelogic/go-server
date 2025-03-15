package httprouter

import (
	"context"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/httprouter"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Prefix     string   `kong:"-"`
	Origin     string   `default:"*" help:"CORS origin"`
	Middleware []string `default:"" help:"Middleware to apply to all routes"`
}

var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	return httprouter.New(ctx, c.Prefix, c.Origin, c.Middleware...)
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return "httprouter"
}

func (c Config) Description() string {
	return "HTTP request router"
}
