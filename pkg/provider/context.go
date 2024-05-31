package provider

import (
	"context"

	"github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ProviderContextKey int

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ ProviderContextKey = iota
	contextLabel
	contextLogger
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithLabel returns a context with the given label
func WithLabel(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, contextLabel, label)
}

// Label returns the label from the context, or zero value if not defined
func Label(ctx context.Context) string {
	if value, ok := ctx.Value(contextLabel).(string); ok {
		return value
	} else {
		return ""
	}
}

// WithLogger returns a context with the given logger
func WithLogger(ctx context.Context, logger server.Logger) context.Context {
	return context.WithValue(ctx, contextLogger, logger)
}

// Logger returns the logger from the context, or nil if not defined
func Logger(ctx context.Context) server.Logger {
	if value, ok := ctx.Value(contextLogger).(server.Logger); ok {
		return value
	} else {
		return nil
	}
}
