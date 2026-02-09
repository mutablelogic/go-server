package server

import (
	"context"
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// LOGGING AND METRICS

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

// Middleware defines methods for HTTP middleware
type Middleware interface {
	HTTPHandlerFunc(http.HandlerFunc) http.HandlerFunc
}
