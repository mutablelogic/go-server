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
	Run(context.Context) error
}

// Router represents a router to which you can add requests
type Router interface {
	// Add a handler to the router, with the given host, path
	// and methods. The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandler(ctx context.Context, hostpath string, handler http.HandlerFunc, methods ...string)

	// Add a handler to the router, with the given host, regular expression
	// path and methods.The context is used to pass additional
	// parameters to the handler. If no methods are provided, then
	// all methods are allowed.
	AddHandlerRe(ctx context.Context, host string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string)
}
