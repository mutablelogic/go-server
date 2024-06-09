package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// route handler, with associated host, prefix and path
type route struct {
	label   string
	host    string
	prefix  string
	path    string
	re      *regexp.Regexp
	handler http.HandlerFunc
	methods []string
	scopes  []string
}

// matchedRoute is a route which has been matched by the router
type matchedRoute struct {
	// Requested Method
	method string

	// Requested Host
	host string

	// Requested Path
	request string

	// Computed	parameters from the path
	parameters []string

	// The route that has been matched
	*route

	// TODO: Whether the result was from the cache
	cached bool
}

var _ server.Route = (*route)(nil)
var _ server.Route = (*matchedRoute)(nil)
var _ Route = (*route)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRoute(ctx context.Context, host, prefix string, methods ...string) *route {
	route := new(route)
	route.label = provider.Label(ctx)
	route.host = host
	route.prefix = prefix
	route.methods = methods
	return route
}

func NewRouteWithPath(ctx context.Context, host, prefix, path string, methods ...string) *route {
	route := NewRoute(ctx, host, prefix, methods...)
	route.path = path

	fmt.Println("new router with path", route)

	return route
}

func NewRouteWithRegexp(ctx context.Context, host, prefix string, path *regexp.Regexp, methods ...string) *route {
	route := NewRoute(ctx, host, prefix, methods...)
	route.re = path
	return route
}

func NewMatchedRoute(route *route, method, host, request string, params ...string) *matchedRoute {
	matched := new(matchedRoute)

	// These parameters are used for caching
	matched.host = host
	matched.method = method
	matched.request = request
	matched.parameters = params

	// The cache stores this route
	matched.route = route

	// Return the matched route
	return matched
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *route) String() string {
	str := "<route"
	if r.label != "" {
		str += fmt.Sprintf(" label=%q", r.label)
	}
	if r.host != "" {
		str += fmt.Sprintf(" host=%q", r.host)
	}
	if r.prefix != "" {
		str += fmt.Sprintf(" prefix=%q", r.prefix)
	}
	if r.path != "" {
		str += fmt.Sprintf(" path=%q", r.path)
	}
	if r.re != nil {
		str += fmt.Sprintf(" re=%q", r.re)
	}
	if r.methods != nil {
		str += fmt.Sprintf(" methods=%v", r.methods)
	}
	if r.scopes != nil {
		str += fmt.Sprintf(" scopes=%v", r.scopes)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *route) Label() string {
	return r.label
}

func (r *route) Host() string {
	return r.host
}

func (r *route) Prefix() string {
	return r.prefix
}

func (r *route) Path() string {
	return r.path
}

func (r *route) Parameters() []string {
	return nil
}

func (r *route) Methods() []string {
	return r.methods
}

func (r *route) Scopes() []string {
	return r.scopes
}

func (r *matchedRoute) Parameters() []string {
	return r.parameters
}

func (r *matchedRoute) Path() string {
	return r.request
}

func (r *route) SetScope(scope ...string) Route {
	for _, s := range scope {
		if !slices.Contains(r.scopes, s) {
			r.scopes = append(r.scopes, s)
		}
	}
	return r
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Match the handler to the path, returning the parameters
// If no match then return nil
func (r *route) MatchRe(path string) []string {
	if r.re == nil {
		return nil
	} else if params := r.re.FindStringSubmatch(path); params != nil {
		return params[1:]
	} else {
		return nil
	}
}

// Match the handler to the path, returning true if it matches
func (r *route) MatchPath(path string) bool {
	if r.path == "" {
		return false
	}
	return strings.HasPrefix(path, r.path)
}

// Match the handler to the method, returning true if it matches
// any of the methods
func (r *route) MatchMethod(method string) bool {
	// Any method is allowed
	if len(r.methods) == 0 {
		return true
	}
	// Specific methods are allowed
	return slices.Contains(r.methods, method)
}
