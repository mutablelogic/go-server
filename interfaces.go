package server

import (
	"context"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Event interface {
	Context() context.Context
	Key() any
	Value() any
	Error() error
}

// EventSource is the interface for subscribing and unsubscribing from
// events
type EventSource interface {
	io.Closer
	Emit(Event) bool
	Sub() <-chan Event
	Unsub(<-chan Event)
}

type EventReceiver interface {
	Rcv(context.Context, func(Event) error, ...EventSource) error
}
