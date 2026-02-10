package httprouter

import (
	"net/http"

	// Packages
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// MiddlewareProvider is implemented by resource instances that supply HTTP
// middleware. The router pulls the function during its own [Apply] and
// appends it to the middleware chain in the order the instances appear in
// the router's middleware attribute.
type MiddlewareProvider interface {
	MiddlewareFunc() HTTPMiddlewareFunc
}

// HandlerProvider is implemented by resource instances that supply an HTTP
// handler for a specific path. The router pulls the handler during its own
// [Apply] and registers it via [Router.RegisterFunc].
type HandlerProvider interface {
	// HandlerPath returns the route path relative to the router prefix
	// (e.g. "resource", "resource/{id}").
	HandlerPath() string

	// HandlerFunc returns the HTTP handler function.
	HandlerFunc() http.HandlerFunc

	// HandlerMiddleware reports whether the router's middleware chain
	// should wrap this handler.
	HandlerMiddleware() bool

	// HandlerSpec returns the OpenAPI path-item description for this
	// handler, or nil if no spec is provided.
	HandlerSpec() *openapi.PathItem
}
