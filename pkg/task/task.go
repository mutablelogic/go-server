package task

import (
	"context"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	event "github.com/mutablelogic/go-server/pkg/event"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Task is a basic task type, which provides a single long-running task
// until cancel is called. It can also
type Task struct {
	event.Source
}

// Compile time check
var _ iface.Task = (*Task)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new default task
func NewTask(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	return new(Task), nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run will block the task run until the context is cancelled or deadline exceeded,
// and then return the reason for cancellation.
func (t *Task) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Task) String() string {
	str := "<task"
	if t.Source.Len() > 0 {
		str += " " + t.Source.String()
	}
	return str + ">"
}
