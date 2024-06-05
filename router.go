package server

import (
	"context"
	"net/http"
	"regexp"
)

// Router represents a router to which you can add requests
type Router interface {
	Task

	// Add a handler to the router, with the given path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandler(context.Context, string, http.Handler, ...string) Route

	// Add a handler function to the router, with the given path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerFunc(context.Context, string, http.HandlerFunc, ...string) Route

	// Add a handler to the router, with the given regular expression
	// path and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerRe(context.Context, *regexp.Regexp, http.Handler, ...string) Route

	// Add a handler function to the router, with the given regular expression
	// path and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerFuncRe(context.Context, *regexp.Regexp, http.HandlerFunc, ...string) Route
}

// Middleware represents an interceptor for HTTP requests
type Middleware interface {
	Task

	// Wrap a handler function
	Wrap(context.Context, http.HandlerFunc) http.HandlerFunc
}

// ServiceEndpoints represents a set of http service endpoints to route requests to
type ServiceEndpoints interface {
	Task

	// Add the endpoints to the router, with the given middleware
	AddEndpoints(context.Context, Router)
}

// Route maps a request to a route handler, with associated host, prefix and path
type Route interface {
	// Label
	Label() string

	// Matched host
	Host() string

	// Matched prefix
	Prefix() string

	// Matching Path
	Path() string

	// Matched parameters
	Parameters() []string

	// Matched methods
	Methods() []string

	// Scopes for authorization
	Scopes() []string
}
