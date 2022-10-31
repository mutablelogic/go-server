package server

import (
	"context"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Event encapsulates a transmitted message with key, value pair. It can
// also encapsulate an error.
type Event interface {
	Context() context.Context // Returns the context of the emitted event from the source
	Key() any                 // Returns the key (or "type") of the event. Returns nil if an error event
	Value() any               // Returns the value of the event. Returns an error if an error event
	Error() error             // Returns the error or nil if not an error event

	// Emit an event on a channel. If the channel is buffered, returns false if the
	// event could not be sent (buffered channel is full). Blocks if an unbuferred channel.
	// Returns false if an error occured.
	Emit(chan<- Event) bool
}

// EventSource is the interface for subscribing and unsubscribing from
// events
type EventSource interface {
	Sub() <-chan Event
	Unsub(<-chan Event)
}

// EventReceiver receives and processes events from one or more
// sources
type EventReceiver interface {
	Rcv(context.Context, func(Event) error, ...EventSource) error
}

// Task is a long-running task which can be a source of events and errors
type Task interface {
	EventSource

	// Run the task until the context is cancelled, and return any errors
	Run(context.Context) error
}

// Plugin creates a task from a configuration
type Plugin interface {
	Name() string                                // Return the name of the task. This should be unique amongst all registered plugins
	Label() string                               // Return the label for the task. This should be unique amongst all plugins with the same name
	New(context.Context, Provider) (Task, error) // Create a new task with provider of other tasks
}

// Provider runs many tasks simultaneously. It subscribes to events from the tasks
// and emits them on its own event channel.
type Provider interface {
	Task
	//plugin.Log
}
