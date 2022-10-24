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

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ctx context.Context, key, value any) Event {
	if key == nil {
		return nil
	}
	return &event{ctx, key, value}
}

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
