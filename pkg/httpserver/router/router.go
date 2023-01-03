package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"sync"

	// Package imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type router struct {
	task.Task
	sync.RWMutex

	label  string
	prefix map[string]Gateway
	routes []*route
}

type Router interface {
	plugin.Router
	AddHandlerEx(string, *regexp.Regexp, http.HandlerFunc, ...string) *route
}

var _ Router = (*router)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new router task, and register routes from gateways
func NewWithPlugin(p Plugin, label string, routes map[string]plugin.Gateway) (*router, error) {
	router := new(router)
	router.prefix = make(map[string]Gateway, len(routes)+1)
	router.label = label

	// If prefix is defined, then register handlers
	parent := context.Background()
	if prefix := p.Prefix(); prefix != "" {
		prefix = normalizePath(prefix, false)
		router.prefix[prefix] = Gateway{
			Label:       router.Label(),
			Description: router.Description(),
		}
		router.RegisterHandlers(ctx.WithPrefix(parent, prefix), router)
	}

	// Register additional routes
	for prefix, gateway := range routes {
		prefix = normalizePath(prefix, false)
		if _, exists := router.prefix[prefix]; exists {
			return nil, ErrDuplicateEntry.Withf("Duplicate prefix %q", prefix)
		}
		router.prefix[prefix] = Gateway{
			Label:       gateway.Label(),
			Description: gateway.Description(),
		}
		gateway.RegisterHandlers(ctx.WithPrefix(parent, prefix), router)
	}

	// Return success
	return router, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (router *router) String() string {
	str := "<httpserver-router"
	if label := router.Label(); label != "" {
		str += fmt.Sprintf(" label=%q", label)
	}
	if prefixes := router.Prefixes(); len(prefixes) > 0 {
		str += fmt.Sprintf(" prefixes=%q", prefixes)
	}
	for _, route := range router.routes {
		str += fmt.Sprint(" ", route)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of the router
func (router *router) Label() string {
	return router.label
}

// Description returns the label of the router
func (router *router) Description() string {
	return "Routes HTTP requests to services and handlers"
}

// Prefixes returns the prefixes recognised by the router
func (router *router) Prefixes() []string {
	prefixes := make([]string, 0, len(router.prefix))
	router.RLock()
	defer router.RUnlock()

	for prefix := range router.prefix {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)

	return prefixes
}

// AddHandler adds a handler to the router, with the context passing the
// prefix and authorization scopes.
func (router *router) AddHandler(parent context.Context, path *regexp.Regexp, fn http.HandlerFunc, methods ...string) {
	route := router.AddHandlerEx(ctx.Prefix(parent), path, fn, methods...)
	if route == nil {
		panic("nil route")
	}

	// Add scopes and description to route
	if scopes := ctx.Scope(parent); len(scopes) > 0 {
		route.scopes = scopes
	}
	if description := ctx.Description(parent); description != "" {
		route.description = description
	}
}

// AddHandlerEx adds a handler to the router, for a specific host/prefix and http methods supported.
// If path argument is nil, then any path under the prefix will match. If the path contains
// a regular expression, then a match is made and any matched parameters of the regular
// expression can be retrieved from the request context.
func (router *router) AddHandlerEx(prefix string, path *regexp.Regexp, fn http.HandlerFunc, methods ...string) *route {
	route := NewRoute(prefix, path, fn, methods...)
	// The priority is either 0 for default routes (where path is nil) or the number of routes, so that
	// handlers are called in the order they are added
	if path != nil {
		route.priority = len(router.routes)
	}

	// Append the route to the list of routes
	router.routes = append(router.routes, route)

	// Sort routes by prefix length, longest first, and then by priority
	sort.Slice(router.routes, func(i, j int) bool {
		if len(router.routes[i].prefix) < len(router.routes[j].prefix) {
			return false
		}
		if len(router.routes[i].prefix) == len(router.routes[j].prefix) && router.routes[i].priority < router.routes[j].priority {
			return false
		}
		return true
	})

	// Return the route
	return route
}

// MatchPath calls the provided function for each route that matches the request
// host and path. Will bail out if true is returned from the function
func (router *router) MatchPath(req *http.Request, fn func(*route, string, []string) bool) {
	for _, route := range router.routes {
		if route.MatchesHost(req.Host) {
			if params, rel, ok := route.MatchesPath(req.URL.Path); ok {
				if fn(route, rel, params) {
					return
				}
			}
		}
	}
}

// ServeHTTP implements the http.Handler interface
func (router *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var matchedPath, matchedMethod bool
	router.MatchPath(req, func(route *route, path string, params []string) bool {
		matchedPath = true
		if route.MatchesMethod(req.Method) {
			matchedMethod = true
			route.fn(w, req.Clone(ctx.WithPrefixPathParams(req.Context(), route.prefix, path, params)))
			// TODO: Cache the route
			return true
		}
		return false
	})

	// Deal with 404 and 405 errors
	if matchedPath && !matchedMethod {
		util.ServeError(w, http.StatusMethodNotAllowed)
	} else if !matchedPath {
		util.ServeError(w, http.StatusNotFound)
	}
}
