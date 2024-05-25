package router

import (
	"context"
	"time"

	"github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type RouterContextKey int

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ RouterContextKey = iota
	contextKey
	contextPrefix
	contextParams
	contextMiddleware
	contextTime
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

// WithTime returns a context with the given time
func WithTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, contextTime, t)
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

// WithMiddleware returns a context with the given middleware
func WithMiddleware(ctx context.Context, middleware ...server.Middleware) context.Context {
	if len(middleware) == 0 {
		return ctx
	}
	return context.WithValue(ctx, contextMiddleware, append(Middleware(ctx), middleware...))
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

// Middleware returns a set of middleware from the context, or zero value if not defined
func Middleware(ctx context.Context) []server.Middleware {
	if value, ok := ctx.Value(contextMiddleware).([]server.Middleware); ok {
		return value
	} else {
		return nil
	}
}

// Time returns the stored time value or the zero value if not defined
func Time(ctx context.Context) time.Time {
	if value, ok := ctx.Value(contextTime).(time.Time); ok {
		return value
	} else {
		return time.Time{}
	}
}