package server

import (
	"context"
	"io/fs"
	"net/http"

	// OpenAPI specification
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
)

///////////////////////////////////////////////////////////////////////////////
// LOGGER

// Logger defines methods for logging messages and structured data.
// It can also act as HTTP middleware for request logging.
type Logger interface {
	// Debug logs a debugging message.
	Debug(context.Context, ...any)

	// Debugf logs a formatted debugging message.
	Debugf(context.Context, string, ...any)

	// Print logs an informational message.
	Print(context.Context, ...any)

	// Printf logs a formatted informational message.
	Printf(context.Context, string, ...any)

	// With returns a new Logger that includes the given key-value pairs
	// in its structured log output.
	With(...any) Logger
}

///////////////////////////////////////////////////////////////////////////////
// ROUTER

// HTTPRouter defines methods for a http router
type HTTPRouter interface {
	http.Handler

	// Spec returns the OpenAPI specification for this router.
	Spec() *openapi.Spec
}

///////////////////////////////////////////////////////////////////////////////
// SERVER

// HTTPServer defines methods for an HTTP server instance.
// The router uses this interface to populate the OpenAPI spec's
// servers list.
type HTTPServer interface {
	// Spec returns the OpenAPI server entry for this instance,
	// or nil if not yet available.
	Spec() *openapi.Server
}

///////////////////////////////////////////////////////////////////////////////
// HANDLER

// HTTPHandler defines methods for HTTP handlers
type HTTPHandler interface {
	// HandlerPath returns the route path relative to the router prefix
	// (e.g. "resource", "resource/{id}").
	HandlerPath() string

	// HandlerFunc returns the HTTP handler function.
	HandlerFunc() http.HandlerFunc

	// Spec returns the OpenAPI path-item description for this
	// handler, or nil if no spec is provided.
	Spec() *openapi.PathItem
}

///////////////////////////////////////////////////////////////////////////////
// FILE SERVER

// HTTPFileServer defines methods for static file serving handlers.
// When the router detects a handler that implements this interface, it
// uses [Router.RegisterFS] instead of [Router.RegisterFunc] so that
// the router prefix is correctly stripped from the request path.
type HTTPFileServer interface {
	// HandlerPath returns the route path relative to the router prefix.
	HandlerPath() string

	// HandlerFS returns the filesystem to serve.
	HandlerFS() fs.FS

	// Spec returns the OpenAPI path-item description for this
	// handler, or nil if no spec is provided.
	Spec() *openapi.PathItem
}

///////////////////////////////////////////////////////////////////////////////
// AUTH

type HTTPAuth interface {
	HTTPMiddleware

	// Spec returns the OpenAPI security-scheme description for this
	// handler, or nil if no spec is provided.
	// TODO Spec() *openapi.SecurityScheme
}

///////////////////////////////////////////////////////////////////////////////
// MIDDLEWARE

// HTTPMiddleware defines methods for HTTP middleware
type HTTPMiddleware interface {
	WrapFunc(http.HandlerFunc) http.HandlerFunc
}
