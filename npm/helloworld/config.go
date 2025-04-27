package helloworld

import (
	"context"
	"embed"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Router server.HTTPRouter `kong:"-"`
	Prefix string            `help:"Path Prefix"`
}

type task struct{}

var _ server.Plugin = Config{}
var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

//go:embed dist
var fs embed.FS

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	if c.Router == nil {
		return nil, httpresponse.ErrInternalError.Withf("Router not set")
	}

	c.Router.HandleFS(ctx, c.Prefix, fs)

	return new(task), nil
}

////////////////////////////////////////////////////////////////////////////////
// CONFIG

func (c Config) Name() string {
	return "helloworld"
}

func (c Config) Description() string {
	return "Hello World example static content"
}

////////////////////////////////////////////////////////////////////////////////
// TASK

func (t *task) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
