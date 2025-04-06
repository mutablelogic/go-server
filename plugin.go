package server

import (
	"context"
	"net/http"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
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
}

///////////////////////////////////////////////////////////////////////////////
// HTTP MIDDLEWARE

type HTTPMiddleware interface {
	// Return a http handler with middleware as the parent handler
	HandleFunc(http.HandlerFunc) http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// LOGGER

type Logger interface {
	// Emit a debugging message
	Debug(context.Context, ...any)

	// Emit a debugging message with formatting
	Debugf(context.Context, string, ...any)

	// Emit an informational message
	Print(context.Context, ...any)

	// Emit an informational message with formatting
	Printf(context.Context, string, ...any)

	// Append structured data to the log in key-value pairs
	// where the key is a string and the value is any type
	With(...any) Logger
}

///////////////////////////////////////////////////////////////////////////////
// PGPOOL

type PG interface {
	// Return the connection pool
	Conn() pg.PoolConn
}

///////////////////////////////////////////////////////////////////////////////
// PGQUEUE

type PGCallback func(context.Context, any) error

type PGQueue interface {
	Task

	// Return the worker name
	Worker() string

	// Register a ticker with a callback, and return the registered ticker
	RegisterTicker(context.Context, pgschema.TickerMeta, PGCallback) (*pgschema.Ticker, error)

	// Register a queue with a callback, and return the registered queue
	RegisterQueue(context.Context, pgschema.Queue, PGCallback) (*pgschema.Queue, error)

	// Create a task for a queue with a payload and optional delay, and return it
	CreateTask(context.Context, string, any, time.Duration) (*pgschema.Task, error)

	// Delete a ticker by name
	DeleteTicker(context.Context, string) error
}
