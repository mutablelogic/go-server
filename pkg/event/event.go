package event

import (
	"context"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type event struct {
	ctx        context.Context
	key, value any
}

// Compile time check
var _ Event = (*event)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new event with context, key and value. Returns nil if the key
// is nil.
func New(ctx context.Context, key, value any) Event {
	if key == nil {
		return nil
	}
	return &event{ctx, key, value}
}

// Create a new error with context. Returns nil if the err parameter
// is nil.
func Error(ctx context.Context, err error) Event {
	if err == nil {
		return nil
	}
	return &event{ctx, nil, err}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e *event) Context() context.Context {
	return e.ctx
}

func (e *event) Key() any {
	return e.key
}

func (e *event) Value() any {
	return e.value
}

func (e *event) Error() error {
	if e.key == nil {
		return e.value.(error)
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (e *event) Emit(ch chan<- Event) bool {
	if ch == nil {
		return false
	}
	if cap(ch) > 0 {
		select {
		case ch <- e:
			return true
		default:
			return false
		}
	}
	ch <- e
	return true
}
