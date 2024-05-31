package server

import (
	"context"
	"net/http"
	"regexp"
)

// Plugin represents a plugin that can create a task
type Plugin interface {
	// Return the unique name for the plugin
	Name() string

	// Return a description of the plugin
	Description() string

	// Create a task from a plugin
	New() (Task, error)
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
	Task

	// Add a handler to the router, with the given path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandler(context.Context, string, http.Handler, ...string)

	// Add a handler function to the router, with the given path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerFunc(context.Context, string, http.HandlerFunc, ...string)

	// Add a handler to the router, with the given regular expression
	// path and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerRe(context.Context, *regexp.Regexp, http.Handler, ...string)

	// Add a handler function to the router, with the given regular expression
	// path and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerFuncRe(context.Context, *regexp.Regexp, http.HandlerFunc, ...string)
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
