package router

import (
	"context"
	"time"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type RouterContextKey int

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ RouterContextKey = iota
	contextHost
	contextPrefix
	contextParams
	contextMiddleware
	contextTime
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithHost returns a context with the given host
func WithHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, contextHost, host)
}

// WithPrefix returns a context with the given prefix
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, contextPrefix, prefix)
}

// WithHostPrefix returns a context with the given host and prefix
func WithHostPrefix(ctx context.Context, host, prefix string) context.Context {
	return WithHost(WithPrefix(ctx, prefix), host)
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
	if route.Label != "" {
		ctx = provider.WithLabel(ctx, route.Label)
	}
	if route.Host != "" {
		ctx = WithHost(ctx, route.Host)
	}
	if route.Prefix != "" {
		ctx = WithPrefix(ctx, route.Prefix)
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

// Host returns the host from the context, or zero value if not defined
func Host(ctx context.Context) string {
	return str(ctx, contextHost)
}

// Prefix returns the prefix from the context, or zero value if not defined
func Prefix(ctx context.Context) string {
	return str(ctx, contextPrefix)
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func str(ctx context.Context, key RouterContextKey) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}
