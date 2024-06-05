package auth

import (
	"context"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type authContextKey int

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_ authContextKey = iota
	contextName
	contextScope
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithTokenName returns a context with the given auth token name
func WithTokenName(ctx context.Context, token Token) context.Context {
	return context.WithValue(ctx, contextName, token.Name)
}

// WithTokenScope returns a context with the given auth token scope
func WithTokenScope(ctx context.Context, token Token) context.Context {
	return context.WithValue(ctx, contextScope, token.Scope)
}

// WithToken returns a context with the given auth token
func WithToken(ctx context.Context, token Token) context.Context {
	if token.Name != "" {
		ctx = WithTokenName(ctx, token)
	}
	if len(token.Scope) > 0 {
		ctx = WithTokenScope(ctx, token)
	}
	return ctx
}

// TokenName returns the token name from the context
func TokenName(ctx context.Context) string {
	return str(ctx, contextName)
}

// TokenScope returns the token scope from the context, or nil
func TokenScope(ctx context.Context) []string {
	if value, ok := ctx.Value(contextScope).([]string); ok {
		return value
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func str(ctx context.Context, key authContextKey) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}
