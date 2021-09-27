package server

import (
	"context"
	"net"
	"net/http"
	"regexp"
)

/////////////////////////////////////////////////////////////////////
// PROVIDER INTERFACES

// Provider provides services to a module
type Provider interface {
	Logger
	Router
	EventQueue

	// Plugins returns a list of registered plugin names
	Plugins() []string

	// GetPlugin returns a named plugin or nil if not available
	GetPlugin(context.Context, string) Plugin

	// GetConfig populates yaml config
	GetConfig(context.Context, interface{}) error
}

// Plugin provides handlers to server
type Plugin interface {
	// Run plugin background tasks until cancelled
	Run(context.Context, Provider) error
}

// Logger providers a logging interface
type Logger interface {
	Print(context.Context, ...interface{})
	Printf(context.Context, string, ...interface{})
}

// Router allows handlers to be added for serving URL paths
type Router interface {
	AddHandler(context.Context, http.Handler, ...string) error
	AddHandlerFunc(context.Context, http.HandlerFunc, ...string) error
	AddHandlerFuncEx(context.Context, *regexp.Regexp, http.HandlerFunc, ...string) error
}

// Middleware intercepts HTTP requests
type Middleware interface {
	// Add a child handler object which intercepts a handler
	AddMiddleware(context.Context, http.Handler) http.Handler

	// Add a child handler function which intercepts a handler function
	AddMiddlewareFunc(context.Context, http.HandlerFunc) http.HandlerFunc
}

// Queue allows posting events and subscription to events from other plugins
type EventQueue interface {
	Post(context.Context, Event)
	Subscribe(context.Context, chan<- Event) error
}

/////////////////////////////////////////////////////////////////////
// EVENT INTERFACE

type Event interface {
	Name() string
	Value() interface{}
}

/////////////////////////////////////////////////////////////////////
// SERVICE DISCOVERY INTERFACE

type Service interface {
	Instance() string
	Service() string
	Name() string
	Host() string
	Port() uint16
	Zone() string
	Addrs() []net.IP
	Txt() []string

	// Return TXT keys and value for a key
	Keys() []string
	ValueForKey(string) string
}
