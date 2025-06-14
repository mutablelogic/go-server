package config

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	metrics "github.com/mutablelogic/go-server/pkg/metrics"
	handler "github.com/mutablelogic/go-server/pkg/metrics/handler"
	schema "github.com/mutablelogic/go-server/pkg/metrics/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Router server.HTTPRouter `kong:"-"` // HTTP Router
	Path   string            `default:"metrics" help:"Path prefix"`
	Tags   map[string]any    `help:"Tags to add to all metrics"`
}

type task struct {
}

var _ server.Plugin = Config{}
var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	// Create a new metrics manager
	manager, err := metrics.New()
	if err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		handler.RegisterMetrics(ctx, c.Router, c.Path, manager)
	}

	// Return the task
	return &task{}, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (task) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return schema.SchemaName
}

func (c Config) Description() string {
	return "Prometheus metrics"
}
