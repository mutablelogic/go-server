package server

import "context"

// Plugin represents a plugin that can create a task
type Plugin interface {
	Name() string
	Description() string
	New(context.Context) (Task, error)
}

// Task represents a task that can be run
type Task interface {
	Run(context.Context) error
}
