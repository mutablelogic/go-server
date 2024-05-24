package router

import "context"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type RouterContextKey string

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	contextKey    RouterContextKey = "key"
	contextPrefix RouterContextKey = "prefix"
	contextParams RouterContextKey = "params"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithPrefix returns a context with the given prefix
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, contextPrefix, prefix)
}

// WithKey returns a context with the given key value
func WithKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, contextKey, key)
}

// WithRoute returns a context with various route parameters
func WithRoute(ctx context.Context, route *Route) context.Context {
	if route == nil {
		return ctx
	}
	if route.Prefix != "" {
		ctx = WithPrefix(ctx, route.Prefix)
	}
	if route.Key != "" {
		ctx = WithKey(ctx, route.Key)
	}
	if len(route.Parameters) > 0 {
		ctx = context.WithValue(ctx, contextParams, route.Parameters)
	}
	return ctx
}

// Prefix returns the prefix from the context, or zero value if not defined
func Prefix(ctx context.Context) string {
	if value, ok := ctx.Value(contextPrefix).(string); ok {
		return value
	} else {
		return ""
	}
}

// Key returns the key from the context, or zero value if not defined
func Key(ctx context.Context) string {
	if value, ok := ctx.Value(contextKey).(string); ok {
		return value
	} else {
		return ""
	}
}

// Params returns the path parameters from the context, or zero value if not defined
func Params(ctx context.Context) []string {
	if value, ok := ctx.Value(contextParams).([]string); ok {
		return value
	} else {
		return nil
	}
}
