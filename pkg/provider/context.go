package provider

import (
	"context"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ProviderContextKey int

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ ProviderContextKey = iota
	contextLabel
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
