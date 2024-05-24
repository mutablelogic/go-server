package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
}

type router struct {
	// Map host to a request router. If host key is empty, then it's a default router
	// where the host is not matched
	host map[string]*reqrouter
}

// represents a set of handlers to be considered for a request
type reqhandlers []*reqhandler

// Ensure interfaces is implemented
var _ http.Handler = (*router)(nil)
var _ server.Plugin = Config{}
var _ server.Task = (*router)(nil)
var _ server.Router = (*router)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "router"
	defaultCap  = 10
	pathSep     = "/"
	hostSep     = "."
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "router for http requests"
}

// Create a new router from the configuration
func (c Config) New(context.Context) (server.Task, error) {
	r := new(router)
	r.host = make(map[string]*reqrouter, defaultCap)

	// Return success
	return r, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Implement the http.Handler interface to route requests
func (router *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, code := router.Match(canonicalHost(r.Host), r.Method, r.URL.Path)

	// Set the path
	//r.URL.Path = route.Path

	// Create a new context
	ctx := WithRoute(r.Context(), route)

	// Create a new request
	r = r.Clone(ctx)

	// TODO: Cache the route if not already cached

	switch code {
	case http.StatusPermanentRedirect:
		http.Redirect(w, r, route.Path, int(code))
	case http.StatusNotFound:
		httpresponse.Error(w, code, "Not found: ", r.URL.Path)
	case http.StatusMethodNotAllowed:
		httpresponse.Error(w, code, "Method not allowed: ", r.Method)
	case http.StatusOK:
		route.Handler(w, r)
	default:
		httpresponse.Error(w, http.StatusInternalServerError, "Internal error: ", fmt.Sprint(code))
	}
}

// Run the router until the context is cancelled
func (router *router) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (router *router) AddHandler(ctx context.Context, hostpath string, handler http.HandlerFunc, methods ...string) {
	// When hostpath is empty, then it's a default handler for all hosts
	if hostpath == "" {
		hostpath = "/"
	}

	// Split into host and path
	parts := strings.SplitN(hostpath, pathSep, 2)
	if len(parts) == 1 {
		parts = append(parts, "")
	}

	// Create a new request router for the host
	key := canonicalHost(parts[0])
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter(key)
	}

	// Add the handler
	router.host[key].AddHandler(ctx, canonicalPrefix(ctx), pathSep+parts[1], handler, methods...)
}

func (router *router) AddHandlerRe(ctx context.Context, host string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) {
	// Create a new request router for the host
	key := canonicalHost(host)
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter(key)
	}
	router.host[key].AddHandlerRe(ctx, canonicalPrefix(ctx), path, handler, methods...)
}

// Match handlers for a given method, host and path, Returns the match
// and the status code, which can be 200, 308, 404 or 405. If the
// status code is 308, then the path is the redirect path.
func (router *router) Match(host, method, path string) (*Route, int) {
	var results reqhandlers

	// Check for host and path
	host = canonicalHost(host)
	for key, r := range router.host {
		if key != "" && !strings.HasSuffix(host, key) {
			continue
		}
		if handlers, redirect := r.matchHandlers(path); len(handlers) > 0 {
			results = append(results, handlers...)
		} else if redirect != "" {
			// Bail out to redirect
			return &Route{
				Path: redirect,
			}, http.StatusPermanentRedirect
		}
	}

	// Bail out if no results
	if len(results) == 0 {
		return nil, http.StatusNotFound
	}

	// Sort results by prefix and path length, with longest first
	sort.Sort(results)

	// Match method
	for _, r := range results {
		// Return the first method which matches
		if !matchMethod(r, method) {
			continue
		}
		// Determine the path
		path := strings.TrimPrefix(path, r.Prefix)
		if path == "" || !strings.HasPrefix(path, pathSep) {
			path = pathSep + path
		}
		// Return the route
		return &Route{
			Key:        r.Key,
			Host:       r.Host,
			Prefix:     r.Prefix,
			Path:       path,
			Parameters: matchHandlerRe(r, path),
			Handler:    r.Handler,
		}, http.StatusOK
	}

	// We had a match but not for the method
	return nil, http.StatusMethodNotAllowed
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return prefix from context, always starts with a '/'
// and never ends with a '/'
func canonicalPrefix(ctx context.Context) string {
	prefix := Prefix(ctx)
	if prefix == "" {
		return pathSep
	}
	return "/" + strings.Trim(prefix, pathSep)
}

// Return host, always starts with a '.' and never ends with a '.'
// or returns empty string if host is empty
func canonicalHost(host string) string {
	if host == "" {
		return ""
	}
	return strings.ToLower(hostSep + strings.Trim(host, hostSep))
}

func (r reqhandlers) Len() int {
	return len(r)
}

func (r reqhandlers) Less(i, j int) bool {
	leni := len(r[i].Prefix) + len(r[i].Path)
	lenj := len(r[j].Prefix) + len(r[j].Path)
	return leni > lenj
}

func (r reqhandlers) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
