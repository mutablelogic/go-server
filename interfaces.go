package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// CMD

// Cmd provides access to the runtime context that is set up by the main entry
// point and passed to command Run methods.
type Cmd interface {
	// Name returns the executable name.
	Name() string

	// Description returns the application description string.
	Description() string

	// Version returns the application version string.
	Version() string

	// Context returns the lifecycle context (cancelled on SIGINT).
	Context() context.Context

	// Logger returns the structured logger.
	Logger() *slog.Logger

	// Tracer returns the OpenTelemetry tracer, or nil if OTel is not configured.
	Tracer() trace.Tracer

	// ClientEndpoint returns the HTTP endpoint URL and client options derived
	// from the global HTTP flags.
	ClientEndpoint() (string, []client.ClientOpt, error)

	// Get retrieves a default value by key. Returns nil if the key does not exist.
	Get(string) any

	// GetString retrieves a default string value by key. Returns empty string if the key
	// does not exist or the value is not a string.
	GetString(string) string

	// Set stores a default value by key and persists the store to disk.
	// Pass nil to remove a key.
	Set(string, any) error

	// Keys returns all keys in the store.
	Keys() []string

	// IsTerm reports whether stderr is an interactive terminal.
	IsTerm() bool

	// IsDebug reports whether debug logging is enabled.
	IsDebug() bool

	// HTTPAddr returns the HTTP listen/connect address.
	HTTPAddr() string

	// HTTPPrefix returns the HTTP path prefix.
	HTTPPrefix() string

	// HTTPTimeout returns the HTTP read/write timeout.
	HTTPTimeout() time.Duration
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
// MIDDLEWARE

// HTTPMiddleware defines methods for HTTP middleware
type HTTPMiddleware interface {
	WrapFunc(http.HandlerFunc) http.HandlerFunc
}
