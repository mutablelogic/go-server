package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	fcgi "github.com/mutablelogic/go-server/pkg/httpserver/fcgi"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Services ServiceConfig `hcl:"services"`
}

type ServiceConfig map[string]struct {
	Service    server.ServiceEndpoints `hcl:"service"`
	Middleware []server.Middleware     `hcl:"middleware"`
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
func (c Config) New() (server.Task, error) {
	r := new(router)
	r.host = make(map[string]*reqrouter, defaultCap)

	// Add services
	for key, service := range c.Services {
		parts := strings.SplitN(key, pathSep, 2)
		if len(parts) == 1 {
			// Could be interpreted as a host if there is a dot in it, or else
			// we assume it's a path
			if strings.Contains(parts[0], hostSep) {
				parts = append(parts, pathSep)
			} else {
				parts, parts[0] = append(parts, parts[0]), ""
			}
		}
		r.addServiceEndpoints(parts[0], parts[1], service.Service, service.Middleware...)
	}

	// Return success
	return r, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Implement the http.Handler interface to route requests
func (router *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := WithTime(r.Context(), time.Now())

	// Process FastCGI environment - remove the REQUEST_PREFIX from the request path
	path := r.URL.Path
	if env := fcgi.ProcessEnv(r); len(env) > 0 {
		if prefix, exists := env["REQUEST_PREFIX"]; exists {
			path = strings.TrimPrefix(path, canonicalPrefix(prefix))
		}
	}

	// Match the route
	route, code := router.Match(canonicalHost(r.Host), r.Method, path)

	// TODO: Cache the route if not already cached

	// Close the body after return
	defer r.Body.Close()

	// Switch on the status code
	switch code {
	case http.StatusPermanentRedirect:
		// TODO: Change this to a JSON redirect - currently returns HTML
		http.Redirect(w, r, r.URL.Path+pathSep, int(code))
	case http.StatusNotFound:
		httpresponse.Error(w, code, "not found:", r.URL.Path)
	case http.StatusMethodNotAllowed:
		httpresponse.Error(w, code, "method not allowed:", r.Method)
	case http.StatusOK:
		r = r.Clone(WithRoute(ctx, route))
		r.URL.Path = route.Path
		route.Handler(w, r)
	default:
		httpresponse.Error(w, http.StatusInternalServerError, "Internal error", fmt.Sprint(code))
	}
}

// Return the label for the task
func (router *router) Label() string {
	// TODO
	return defaultName
}

// Run the router until the context is cancelled
func (router *router) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (router *router) AddHandler(ctx context.Context, path string, handler http.Handler, methods ...string) {
	router.AddHandlerFunc(ctx, path, handler.ServeHTTP, methods...)
}

func (router *router) AddHandlerFunc(ctx context.Context, path string, handler http.HandlerFunc, methods ...string) {
	// Fix the path
	if !strings.HasPrefix(path, pathSep) {
		path = pathSep + path
	}
	// Create a new request router for the host
	key := canonicalHost(Host(ctx))
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter(key)
	}
	router.host[key].AddHandler(ctx, canonicalPrefix(Prefix(ctx)), path, handler, methods...)
}

func (router *router) AddHandlerRe(ctx context.Context, path *regexp.Regexp, handler http.Handler, methods ...string) {
	router.AddHandlerFuncRe(ctx, path, handler.ServeHTTP, methods...)
}

func (router *router) AddHandlerFuncRe(ctx context.Context, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) {
	// Create a new request router for the host
	key := canonicalHost(Host(ctx))
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter(key)
	}
	router.host[key].AddHandlerRe(ctx, canonicalPrefix(Prefix(ctx)), path, handler.ServeHTTP, methods...)
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
			Label:      r.Label,
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

// Add a set of endpoints to the router with a prefix and middleware
func (router *router) addServiceEndpoints(host, prefix string, service server.ServiceEndpoints, middleware ...server.Middleware) {
	// Set the context
	ctx := WithHostPrefix(context.Background(), canonicalHost(host), prefix)
	if len(middleware) > 0 {
		ctx = WithMiddleware(ctx, middleware...)
	}
	if label := service.Label(); label != "" {
		ctx = provider.WithLabel(ctx, label)
	}

	// Call the service to add the endpoints
	service.AddEndpoints(ctx, router)
}

// Return prefix from context, always starts with a '/'
// and never ends with a '/'
func canonicalPrefix(prefix string) string {
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
