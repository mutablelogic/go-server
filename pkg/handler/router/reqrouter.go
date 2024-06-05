package router

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Represents a set of routes which are matched by a host/prefix combination
type reqrouter struct {
	// The host string
	host string

	// Maps prefixes to an array of handlers
	prefix map[string]*reqs
}

// Represents a set of handlers for a prefix
type reqs struct {
	host     string
	prefix   string
	handlers []*route
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a new request router
func newReqRouter(host string) *reqrouter {
	r := new(reqrouter)
	r.host = host
	r.prefix = make(map[string]*reqs, defaultCap)
	return r
}

// Return a new set of requests
func newReqs(host, prefix string) *reqs {
	r := new(reqs)
	r.host = host
	r.prefix = prefix
	r.handlers = make([]*route, 0, defaultCap)
	return r
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (router *reqrouter) AddHandler(ctx context.Context, prefix, path string, handler http.HandlerFunc, methods ...string) *route {
	// Add a path separator to the end of the prefix
	key := prefix
	if prefix != pathSep {
		key = prefix + pathSep
	}
	// Make a new set of requests associated with the prefix
	if _, exists := router.prefix[key]; !exists {
		router.prefix[key] = newReqs(router.host, prefix)
	}

	// Add the handler to the set of requests
	return router.prefix[key].AddHandler(ctx, path, handler, methods...)
}

func (router *reqrouter) AddHandlerRe(ctx context.Context, prefix string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) *route {
	// Add a path separator to the end of the prefix
	key := prefix
	if prefix != pathSep {
		key = prefix + pathSep
	}

	// Make a new set of requests associated with the prefix
	if _, exists := router.prefix[key]; !exists {
		router.prefix[key] = newReqs(router.host, prefix)
	}

	// Add the handler to the set of requests
	return router.prefix[key].AddHandlerRe(ctx, path, handler, methods...)
}

func (router *reqs) AddHandler(ctx context.Context, path string, handler http.HandlerFunc, methods ...string) *route {
	// Add any middleware to the handler, in reverse order
	middleware := Middleware(ctx)
	slices.Reverse(middleware)
	for _, middleware := range middleware {
		handler = middleware.Wrap(ctx, handler)
	}

	// Create the route
	route := NewRouteWithPath(ctx, router.host, router.prefix, path, methods...)

	// Set the route handler
	route.handler = handler

	// Add the handler to the list of handlers
	router.handlers = append(router.handlers, route)

	// Return success
	return route
}

func (router *reqs) AddHandlerRe(ctx context.Context, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) *route {
	// Add any middleware to the handler, in reverse order
	middleware := Middleware(ctx)
	slices.Reverse(middleware)
	for _, middleware := range middleware {
		handler = middleware.Wrap(ctx, handler)
	}

	// Create the route
	route := NewRouteWithRegexp(ctx, router.host, router.prefix, path, methods...)

	// Set the route handler
	route.handler = handler

	// Add the handler to the list of handlers
	router.handlers = append(router.handlers, route)

	// Return success
	return route
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns a set of requests for a path
func (router *reqrouter) matchHandlers(path string) ([]*route, string) {
	// We assume that the path starts with a "/"
	for key, value := range router.prefix {
		// Match the prefix as a path
		if strings.HasPrefix(path, key) {
			if value.prefix != pathSep {
				// Remove the prefix from the path, before matching
				path = strings.TrimPrefix(path, value.prefix)
			}
			return value.matchHandlers(path), ""
		}
		// Match the prefix without a path, and redirect
		if path == value.prefix {
			return nil, value.prefix + pathSep
		}
	}

	// No match
	return nil, ""
}

// Return a list of handlers that match the path
func (router *reqs) matchHandlers(path string) []*route {
	// Assume the path starts with a "/" and the prefix has been removed
	result := make([]*route, 0, defaultCap)
	for _, r := range router.handlers {
		if params := r.MatchRe(path); params != nil {
			result = append(result, r)
		} else if r.MatchPath(path) {
			result = append(result, r)
		}
	}
	return result
}
