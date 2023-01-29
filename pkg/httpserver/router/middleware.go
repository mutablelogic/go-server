package router

import (
	"context"
	"fmt"
	"net/http"

	// Package imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	"github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Middleware is a handler which is called before or after the request
type Middleware struct {
	provider    string
	prefix      string
	description string
	fn          MiddlewareFn
}

// MiddlewareFn returns a new http.HandlerFunc given a "child" http.HandlerFunc
type MiddlewareFn func(http.HandlerFunc) http.HandlerFunc

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (middleware Middleware) String() string {
	str := "<middleware"
	if middleware.provider != "" {
		str += fmt.Sprintf(" provider=%q", middleware.provider)
	}
	if middleware.prefix != "" {
		str += fmt.Sprintf(" prefix=%q", middleware.prefix)
	}
	if middleware.description != "" {
		str += fmt.Sprintf(" description=%q", middleware.description)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Register a middleware handler to the router given unique name
func (router *router) AddMiddleware(parent context.Context, fn func(http.HandlerFunc) http.HandlerFunc) error {
	// Get the name of the middleware and check it is valid
	prefix := ctx.Prefix(parent)
	provider := ctx.NameLabel(parent)
	if prefix == "" {
		return ErrBadParameter.Withf("%s: missing prefix", provider)
	} else if !types.IsIdentifier(prefix) {
		return ErrBadParameter.Withf("%s: invalid prefix: %q", provider, prefix)
	} else if _, exists := router.middleware[prefix]; exists {
		return ErrDuplicateEntry.Withf("%s: duplicate prefix: %q", provider, prefix)
	}

	// Append the middleware
	router.middleware[prefix] = Middleware{
		provider:    provider,
		prefix:      prefix,
		description: ctx.Description(parent),
		fn:          fn,
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (router *router) wrap(handler http.HandlerFunc, middleware []string) (http.HandlerFunc, error) {
	// Wrap the handler function. This is done in reverse order so the first/last middleware called
	// is the first in the list of middleware
	for i := len(middleware); i > 0; i-- {
		prefix := middleware[i-1]
		if fn, exists := router.middleware[prefix]; !exists {
			return nil, ErrNotFound.Withf("middleware not found: %q", prefix)
		} else {
			handler = fn.fn(handler)
		}
	}

	// Return the wrapped handler function
	return handler, nil
}
