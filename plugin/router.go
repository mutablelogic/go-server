package plugin

import (
	"context"
	"net/http"
	"regexp"
)

// Router is a task which maps paths to handlers
type Router interface {
	http.Handler

	// Add a host/prefix/path mapping to a handler for one or more HTTP methods
	// which match the regular expression. If the regular expression is nil,
	// then any path is matched. The methods which are supported by the
	// handler are determined are provided by the final argument. If no
	// methods are provided, then the GET method is assumed.
	AddHandler(context.Context, *regexp.Regexp, http.HandlerFunc, ...string)

	// Register a middleware handler to the router given unique name
	AddMiddleware(context.Context, func(http.HandlerFunc) http.HandlerFunc) error
}
