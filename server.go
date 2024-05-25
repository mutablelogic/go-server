package server

import (
	"context"
	"net/http"
	"regexp"
)

// Plugin represents a plugin that can create a task
type Plugin interface {
	Name() string
	Description() string
	New(context.Context) (Task, error)
}

// Task represents a task that can be run
type Task interface {
	// Return the label for the task
	Label() string

	// Run the task until the context is cancelled and return any errors
	Run(context.Context) error
}

// Router represents a router to which you can add requests
type Router interface {
	// Add a handler to the router, with the given host, path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandler(ctx context.Context, hostpath string, handler http.Handler, methods ...string)

	// Add a handler to the router, with the given host, path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerFunc(ctx context.Context, hostpath string, handler http.HandlerFunc, methods ...string)

	// Add a handler to the router, with the given host, regular expression
	// path and methods.The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerFuncRe(ctx context.Context, host string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string)

	// Add a handler to the router, with the given host, regular expression
	// path and methods.The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerRe(ctx context.Context, host string, path *regexp.Regexp, handler http.Handler, methods ...string)
}

// Middleware represents an interceptor for HTTP requests
type Middleware interface {
	// Wrap a handler function
	Wrap(context.Context, http.HandlerFunc) http.HandlerFunc
}

// Logger interface
type Logger interface {
	// Print logging message
	Print(context.Context, ...any)

	// Print formatted logging message
	Printf(context.Context, string, ...any)
}

// Provider interface
type Provider interface {
	Task
	Logger
}