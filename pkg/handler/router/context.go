package router

import "context"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type RouterContextKey string

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	ContextKeyPrefix RouterContextKey = "prefix"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithPrefix returns a context with the given prefix
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, ContextKeyPrefix, prefix)
}

// Prefix returns the prefix from the context, or zero value if not defined
func Prefix(ctx context.Context) string {
	if value, ok := ctx.Value(ContextKeyPrefix).(string); ok {
		return value
	} else {
		return ""
	}
}
