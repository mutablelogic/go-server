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
	"golang.org/x/exp/maps"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type router struct {
	// Map host to a request router. If host key is empty, then it's a default router
	// where the host is not matched
	host map[string]*reqrouter
}

// represents a set of handlers to be considered for a request
type reqhandlers []*route

// Ensure interfaces is implemented
var _ http.Handler = (*router)(nil)
var _ server.Router = (*router)(nil)
var _ Router = (*router)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new router from the configuration
func New(c Config) (server.Task, error) {
	r := new(router)
	r.host = make(map[string]*reqrouter, defaultCap)

	// Add services
	for key, service := range c.Services {
		r.AddServiceEndpoints(key, service.Service, service.Middleware...)
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
	// and set the host from SERVER_NAME
	path := r.URL.Path
	if env := fcgi.ProcessEnv(r); len(env) > 0 {
		if prefix, exists := env[envRequestPrefix]; exists {
			path = strings.TrimPrefix(path, canonicalPrefix(prefix))
		}
		if serverName, exists := env[envServerName]; exists {
			r.Host = serverName
		}
	}

	// Match the route (TODO: From cache)
	matchedRoute, code := router.Match(r.Method, canonicalHost(r.Host), path)

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
		r = r.Clone(WithRoute(ctx, matchedRoute))
		r.URL.Path = matchedRoute.request
		matchedRoute.route.handler(w, r)
	default:
		httpresponse.Error(w, http.StatusInternalServerError, "Internal error", fmt.Sprint(code))
	}
}

func (router *router) AddHandler(ctx context.Context, path string, handler http.Handler, methods ...string) server.Route {
	return router.AddHandlerFunc(ctx, path, handler.ServeHTTP, methods...)
}

func (router *router) AddHandlerFunc(ctx context.Context, path string, handler http.HandlerFunc, methods ...string) server.Route {
	// Fix the path
	if !strings.HasPrefix(path, pathSep) {
		path = pathSep + path
	}
	// Create a new request router for the host
	key := canonicalHost(Host(ctx))
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter(key)
	}

	// Add the handler to the set of requests
	return router.host[key].AddHandler(ctx, canonicalPrefix(Prefix(ctx)), path, handler, methods...)
}

func (router *router) AddHandlerRe(ctx context.Context, path *regexp.Regexp, handler http.Handler, methods ...string) server.Route {
	return router.AddHandlerFuncRe(ctx, path, handler.ServeHTTP, methods...)
}

func (router *router) AddHandlerFuncRe(ctx context.Context, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) server.Route {
	// Create a new request router for the host
	key := canonicalHost(Host(ctx))
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter(key)
	}

	// Add the handler to the set of requests
	return router.host[key].AddHandlerRe(ctx, canonicalPrefix(Prefix(ctx)), path, handler.ServeHTTP, methods...)
}

// Match handlers for a given method, host and path, Returns the match
// and the status code, which can be 200, 308, 404 or 405. If the
// status code is 308, then the path is the redirect path.
func (router *router) Match(method, host, path string) (*matchedRoute, int) {
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
			return NewMatchedRoute(nil, "", "", redirect), http.StatusPermanentRedirect
		}
	}

	// Bail out if no results
	if len(results) == 0 {
		return nil, http.StatusNotFound
	}

	// If it's an OPTIONS request, then return the route which returns allowed methods
	if method == http.MethodOptions {
		return NewMatchedRoute(newOptionsRoute(results), method, host, path), http.StatusOK
	}

	// Sort results by prefix and path length, with longest first
	sort.Sort(results)

	// Match method
	for _, r := range results {
		// Return the first method which matches
		if !r.MatchMethod(method) {
			continue
		}

		// Determine the path
		path := strings.TrimPrefix(path, r.prefix)
		if path == "" || !strings.HasPrefix(path, pathSep) {
			path = pathSep + path
		}

		// Return the route
		return NewMatchedRoute(r, method, host, path, r.MatchRe(path)...), http.StatusOK
	}

	// We had a match but not for the method
	return nil, http.StatusMethodNotAllowed
}

func (router *router) Scopes() []string {
	scopes := make(map[string]bool)
	for _, r := range router.host {
		for _, h := range r.prefix {
			for _, r := range h.handlers {
				for _, s := range r.scopes {
					scopes[s] = true
				}
			}
		}
	}

	// Gather all scopes
	result := make([]string, 0, len(scopes))
	for scope := range scopes {
		result = append(result, scope)
	}

	// Sort alphabetically
	sort.Strings(result)

	// Return the result
	return result
}

// Add a service endpoint to the router. Returns an error if the prefix is
// already in use.
func (router *router) AddServiceEndpoints(hostprefix string, service server.ServiceEndpoints, middleware ...server.Middleware) {
	parts := strings.SplitN(hostprefix, pathSep, 2)
	if len(parts) == 1 {
		// Could be interpreted as a host if there is a dot in it, or else
		// we assume it's a path
		if strings.Contains(parts[0], hostSep) {
			parts = append(parts, pathSep)
		} else {
			parts, parts[0] = append(parts, parts[0]), ""
		}
	}
	router.addServiceEndpoints(parts[0], parts[1], service, middleware...)
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
	leni := len(r[i].prefix) + len(r[i].path)
	lenj := len(r[j].prefix) + len(r[j].path)
	return leni > lenj
}

func (r reqhandlers) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func newOptionsRoute(handlers reqhandlers) *route {
	methods := make(map[string]bool, len(handlers))
	methods[http.MethodOptions] = true
	for _, h := range handlers {
		for _, m := range h.methods {
			methods[m] = true
		}
	}
	return &route{
		handler: func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Allow", strings.Join(maps.Keys(methods), ", "))
			w.WriteHeader(http.StatusNoContent)
		},
	}
}
