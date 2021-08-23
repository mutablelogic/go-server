package httprouter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	// Modules

	provider "github.com/djthorpe/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type router struct {
	sync.Mutex
	*http.ServeMux
	cache
	routes map[string]*route
}

type route struct {
	cache
	prefix  string
	methods map[string]*routehandlers
}

type routehandlers struct {
	re map[*regexp.Regexp]http.HandlerFunc
	de http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	PathSeparator = string(os.PathSeparator)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *router {
	this := new(router)
	this.ServeMux = http.NewServeMux()
	this.routes = make(map[string]*route)
	return this
}

///////////////////////////////////////////////////////////////////////////////
// ROUTER IMPLEMENTATION

func (this *router) Handler() http.Handler {
	return this.ServeMux
}

func (this *router) AddHandler(ctx context.Context, handler http.Handler, methods ...string) error {
	return this.AddHandlerFunc(ctx, handler.ServeHTTP, methods...)
}

func (this *router) AddHandlerFunc(ctx context.Context, handler http.HandlerFunc, methods ...string) error {
	return this.AddHandlerFuncEx(ctx, nil, handler, methods...)
}

func (this *router) AddHandlerFuncEx(ctx context.Context, re *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Obtain prefix, set method if not already set
	prefix := prefixForContext(ctx)
	if len(methods) == 0 {
		methods = []string{http.MethodGet}
	}

	// Create a route (which can handle several methods and regular expressions)
	r, exists := this.routes[prefix]
	if !exists {
		r = new(route)
		r.cache = this.cache
		r.prefix = prefix
		r.methods = make(map[string]*routehandlers)
		this.routes[prefix] = r
	}

	// Add to ServeMux
	this.ServeMux.HandleFunc(prefix, r.ServeHTTP)

	// Add handler to route
	return r.AddHandler(re, handler, methods...)
}

func (this *route) AddHandler(re *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	for _, method := range methods {
		if _, exists := this.methods[method]; !exists {
			this.methods[method] = new(routehandlers)
		}
		if re == nil {
			this.methods[method].de = handler
		} else {
			this.methods[method].re[re] = handler
		}
	}

	// Return success
	return nil
}

func (this *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check cache for existing handler
	if handler, params, hits := this.cache.Get(r.Method + r.URL.Path); hits > 0 {
		handler(w, reqWithParams(r, params))
		return
	}

	// Find handler
	handler, exists := this.methods[r.Method]
	if !exists {
		ServeError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method Not Allowed: %q", r.Method))
		return
	}

	// Remove prefix from path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, this.prefix)

	// Find regular expression handler
	for re, handler := range handler.re {
		if args := re.FindStringSubmatch(r.URL.Path); len(args) >= 1 {
			handler(w, reqWithParams(r, args[1:]))
			this.cache.Set(r.Method+r.URL.Path, handler, args[1:])
			return
		}
	}

	// Use default handler
	if handler.de != nil {
		handler.de(w, r)
		this.cache.Set(r.Method+r.URL.Path, handler.de, nil)
		return
	}

	// Handler not found
	ServeError(w, http.StatusNotFound, fmt.Sprintf("Not Found: %q", r.URL.Path))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func reqWithParams(r *http.Request, params []string) *http.Request {
	return r.Clone(provider.ContextWithParams(r.Context(), params))
}

func prefixForContext(ctx context.Context) string {
	prefix := provider.ContextHandlerPrefix(ctx)
	if prefix == "" || prefix == PathSeparator {
		return PathSeparator
	} else {
		return "/" + strings.Trim(prefix, PathSeparator)
	}
}
