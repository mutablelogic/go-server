package event

import (
	"context"
	"fmt"
	"strconv"

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

// Create an informational event
func Infof(ctx context.Context, key any, format string, a ...any) Event {
	if format == "" || key == nil {
		return nil
	}
	return &event{ctx, key, fmt.Sprintf(format, a...)}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return the event as a string
func (e *event) String() string {
	str := "<event"
	if e.key != nil {
		str += fmt.Sprint(" key=", toString(e.key))
		if e.value != nil {
			str += fmt.Sprint(" value=", toString(e.value))
		}
	} else {
		str += fmt.Sprint(" error=", e.value)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the event context, which can be nil if there is no context
func (e *event) Context() context.Context {
	return e.ctx
}

// Return the event key, which can be nil if the event is an error type
func (e *event) Key() any {
	return e.key
}

// Return the event value, which will be nil if the event is an error type
func (e *event) Value() any {
	if e.key == nil {
		return nil
	} else {
		return e.value
	}
}

// Return the event error value, or nil if the event is not an error type
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

func toString(v any) string {
	switch v := v.(type) {
	case string:
		return strconv.Quote(v)
	default:
		return fmt.Sprint(v)
	}
}
