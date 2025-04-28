package server

import (
	"context"
	"io/fs"
	"net/http"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	authschema "github.com/mutablelogic/go-server/pkg/auth/schema"
	pgschema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Plugin represents a service
type Plugin interface {
	// Return the unique name for the plugin
	Name() string

	// Return a description of the plugin
	Description() string

	// Create a task from a plugin
	New(context.Context) (Task, error)
}

// Task represents a service task
type Task interface {
	// Run a task
	Run(context.Context) error
}

// Provider represents a service provider
type Provider interface {
	Task

	// Load a plugin by name and label
	Load(string, string, func(config Plugin)) error

	// Return a task given a plugin label
	Task(context.Context, string) Task
}

///////////////////////////////////////////////////////////////////////////////
// HTTP ROUTER

type HTTPRouter interface {
	// Return the CORS origin
	Origin() string

	// Register a function to handle a URL path
	HandleFunc(context.Context, string, http.HandlerFunc)

	// Register serving of static files from a filesystem
	HandleFS(context.Context, string, fs.FS)
}

///////////////////////////////////////////////////////////////////////////////
// HTTP MIDDLEWARE

type HTTPMiddleware interface {
	// Return a http handler with middleware as the parent handler
	HandleFunc(http.HandlerFunc) http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// LOGGER

// Logger defines methods for logging messages and structured data.
// It can also act as HTTP middleware for request logging.
type Logger interface {
	HTTPMiddleware

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
// PGPOOL

// PG provides access to a PostgreSQL connection pool.
type PG interface {
	// Conn returns the underlying connection pool object.
	Conn() pg.PoolConn
}

///////////////////////////////////////////////////////////////////////////////
// PGQUEUE

// PGCallback defines the function signature for handling tasks dequeued
// from a PostgreSQL-backed queue.
type PGCallback func(context.Context, any) error

// PGQueue defines methods for interacting with a PostgreSQL-backed task queue.
type PGQueue interface {
	// RegisterTicker registers a periodic task (ticker) with a callback function.
	// It returns the metadata of the registered ticker.
	RegisterTicker(context.Context, pgschema.TickerMeta, PGCallback) (*pgschema.Ticker, error)

	// RegisterQueue registers a task queue with a callback function.
	// It returns the metadata of the registered queue.
	RegisterQueue(context.Context, pgschema.QueueMeta, PGCallback) (*pgschema.Queue, error)

	// CreateTask adds a new task to a specified queue with a payload and optional delay.
	// It returns the metadata of the created task.
	CreateTask(context.Context, string, any, time.Duration) (*pgschema.Task, error)

	// UnmarshalPayload unmarshals a payload into a destination object.
	UnmarshalPayload(dest any, payload any) error
}

///////////////////////////////////////////////////////////////////////////////
// AUTHENTICATION AND AUTHORIZATION

// Auth defines methods for authenticating users based on HTTP requests
// and authorizing them based on required scopes.
type Auth interface {
	// Authenticate attempts to identify and return the user associated with
	// an incoming HTTP request, typically by inspecting headers or tokens.
	Authenticate(*http.Request) (*authschema.User, error)

	// Authorize checks if a given user has the necessary permissions (scopes)
	// to perform an action.
	Authorize(context.Context, *authschema.User, ...string) error
}
