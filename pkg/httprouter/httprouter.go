package httprouter

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	// Packages
	provider "github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type router struct {
	sync.Mutex
	*http.ServeMux
	*cache
	routes map[string]*route
}

type route struct {
	*cache
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
	this.cache = new(cache)
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

	// Modify prefix to include a final slash
	if !strings.HasSuffix(prefix, PathSeparator) {
		prefix += PathSeparator
	}

	// Create a route (which can handle several methods and regular expressions)
	r, exists := this.routes[prefix]
	if !exists {
		r = new(route)
		r.cache = this.cache
		r.prefix = prefix
		r.methods = make(map[string]*routehandlers)
		this.routes[prefix] = r

		// Add to ServeMux
		this.ServeMux.HandleFunc(prefix, r.ServeHTTP)
	}

	// Add handler to route
	return r.AddHandler(re, handler, methods...)
}

func (this *route) AddHandler(re *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	for _, method := range methods {
		if _, exists := this.methods[method]; !exists {
			this.methods[method] = new(routehandlers)
			this.methods[method].re = make(map[*regexp.Regexp]http.HandlerFunc)
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
	// Make cache key
	key := r.Method + r.URL.Path

	// Remove prefix from path
	if !strings.HasPrefix(r.URL.Path, this.prefix) {
		ServeError(w, http.StatusNotFound, ErrBadParameter.With(r.URL.Path).Error())
	} else {
		r.URL.Path = PathSeparator + strings.TrimPrefix(r.URL.Path, this.prefix)
	}

	// Check cache for existing handler
	if handler, params, hits := this.cache.Get(key); hits > 0 {
		handler(w, reqWithParams(r, params))
		return
	}

	// Find handler
	handler, exists := this.methods[r.Method]
	if !exists {
		ServeError(w, http.StatusMethodNotAllowed, ErrBadParameter.With(r.Method).Error())
		return
	}

	// Find regular expression handler
	for re, handler := range handler.re {
		if args := re.FindStringSubmatch(r.URL.Path); args != nil {
			handler(w, reqWithParams(r, args[1:]))
			this.cache.Set(key, handler, args[1:])
			return
		}
	}

	// Use default handler
	if handler.de != nil {
		handler.de(w, r)
		this.cache.Set(key, handler.de, nil)
		return
	}

	// Handler not found
	ServeError(w, http.StatusNotFound, ErrBadParameter.With(r.URL.Path).Error())
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func reqWithParams(r *http.Request, params []string) *http.Request {
	return r.Clone(provider.ContextWithPathParams(r.Context(), r.URL.Path, params))
}

func prefixForContext(ctx context.Context) string {
	prefix := provider.ContextHandlerPrefix(ctx)
	if prefix == "" || prefix == PathSeparator {
		return PathSeparator
	} else {
		return "/" + strings.Trim(prefix, PathSeparator)
	}
}
