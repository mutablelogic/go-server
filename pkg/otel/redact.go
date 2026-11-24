package otel

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// redactions are the path values of a request which are secret, by their
// names, which a handler marks, with RedactPathValue, and the middleware
// leaves out of its log, and its span
type redactions struct {
	mu     sync.Mutex
	values map[string]string // The values, by their names
}

// redactKey is the key of a request's redactions, in its context
type redactKey struct{}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RedactPathValue marks the request's path values, by their names, such as
// a token in /hook/{token}, as secret, so HTTPHandler's log, and span, have
// the path with {token} in the value's place. A handler calls it, of a
// request which HTTPHandler is serving, before it returns. A name without a
// value is ignored, as is a request which HTTPHandler isn't serving.
func RedactPathValue(r *http.Request, names ...string) {
	red, ok := r.Context().Value(redactKey{}).(*redactions)
	if !ok {
		return
	}
	red.mu.Lock()
	defer red.mu.Unlock()
	for _, name := range names {
		if value := r.PathValue(name); value != "" {
			red.values[name] = value
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// withRedactions returns the context, with redactions, which handlers mark
func withRedactions(ctx context.Context) (context.Context, *redactions) {
	red := &redactions{values: make(map[string]string)}
	return context.WithValue(ctx, redactKey{}, red), red
}

// redact returns the text, such as a path, with each value which is marked
// replaced by its name, in braces, and whether there were any
func (red *redactions) redact(text string) (string, bool) {
	red.mu.Lock()
	defer red.mu.Unlock()
	if len(red.values) == 0 {
		return text, false
	}
	for name, value := range red.values {
		text = strings.ReplaceAll(text, value, "{"+name+"}")
	}
	return text, true
}
