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
	contextRequest
	contextMiddleware
	contextScope
	contextMethod
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

// WithScope returns a context with the given scooes
func WithScope(ctx context.Context, scope ...string) context.Context {
	return context.WithValue(ctx, contextScope, scope)
}

// WithMethod returns a context with the given methods
func WithMethod(ctx context.Context, method ...string) context.Context {
	return context.WithValue(ctx, contextMethod, method)
}

// WithMiddleware returns a context with the given middleware
func WithMiddleware(ctx context.Context, middleware ...server.Middleware) context.Context {
	if len(middleware) == 0 {
		return ctx
	}
	return context.WithValue(ctx, contextMiddleware, append(Middleware(ctx), middleware...))
}

// WithRoute returns a context with various route parameters
func WithRoute(ctx context.Context, route *matchedRoute) context.Context {
	if route == nil {
		return ctx
	}
	if route.route != nil {
		if route.label != "" {
			ctx = provider.WithLabel(ctx, route.label)
		}
		if route.host != "" {
			ctx = WithHost(ctx, route.host)
		}
		if route.prefix != "" {
			ctx = WithPrefix(ctx, route.prefix)
		}
		if len(route.scopes) > 0 {
			ctx = WithScope(ctx, route.scopes...)
		}
		if len(route.methods) > 0 {
			ctx = WithMethod(ctx, route.methods...)
		}
	}
	if len(route.parameters) > 0 {
		ctx = context.WithValue(ctx, contextParams, route.parameters)
	}
	if route.request != "" {
		ctx = context.WithValue(ctx, contextRequest, route.request)
	}
	return ctx
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
	return strslice(ctx, contextParams)
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

// Scope returns the stored scope or nil if not defined
func Scope(ctx context.Context) []string {
	return strslice(ctx, contextScope)
}

// Method returns the stored methods or nil if not defined
func Method(ctx context.Context) []string {
	return strslice(ctx, contextMethod)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func str(ctx context.Context, key RouterContextKey) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}

func strslice(ctx context.Context, key RouterContextKey) []string {
	if value, ok := ctx.Value(key).([]string); ok {
		return value
	}
	return nil
}
