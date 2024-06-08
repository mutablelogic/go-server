package router

import (
	// Packages
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Router interface {
	server.Router

	// Match a host, method and path to a handler. Returns the appropriate
	// http status code, which will be 200 on success, 404 or 405 and
	// path parameters extracted from the path.
	Match(host, method, path string) (*matchedRoute, int)

	// Add a service endpoint to the router, with specified host/prefix combinarion.
	AddServiceEndpoints(string, server.ServiceEndpoints, ...server.Middleware)

	// Return all known scopes
	Scopes() []string
}

type Route interface {
	server.Route

	// Set scope
	SetScope(...string) Route
}
