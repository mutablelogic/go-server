package server

import (
	"context"
)

// Plugin represents a plugin that can create a task
type Plugin interface {
	// Return the unique name for the plugin
	Name() string

	// Return a description of the plugin
	Description() string

	// Create a task from a plugin
	New() (Task, error)
}

// Task represents a task that can be run
type Task interface {
	// Return the label for the task
	Label() string

	// Run the task until the context is cancelled and return any errors
	Run(context.Context) error
}

// Logger interface
type Logger interface {
	// Print logging message
	Print(context.Context, ...any)

	// Print formatted logging message
	Printf(context.Context, string, ...any)
}

// Provider interface
type Provider interface {
	Task
	Logger
}
