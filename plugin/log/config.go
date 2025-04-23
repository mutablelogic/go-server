package main

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Debug bool `kong:"-"`
}

var _ server.Plugin = (*Config)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOABALS

const (
	defaultPluginName = "log"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	// Return the task
	return newTaskWith(c), nil
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func Plugin() server.Plugin {
	return Config{}
}

func (c Config) Name() string {
	return defaultPluginName
}

func (c Config) Description() string {
	return "Structured logger"
}
