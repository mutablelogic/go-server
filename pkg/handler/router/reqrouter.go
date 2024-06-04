package router

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/mutablelogic/go-server/pkg/provider"
	// Packages
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Represents a set of routes which are matched by a host/prefix combination
type reqrouter struct {
	// The host string
	host string

	// Maps prefixes to an array of handlers
	prefix map[string]*reqs
}

// Represents a set of handlers for a prefix
type reqs struct {
	host     string
	prefix   string
	handlers []*reqhandler
}

// Represents a handler for a request
type reqhandler struct {
	Label   string           `json:"label,omitempty"`
	Host    string           `json:"host,omitempty"`
	Prefix  string           `json:"prefix,omitempty"`
	Path    string           `json:"path,omitempty"`
	Re      *regexp.Regexp   `json:"re,omitempty"`
	Method  []string         `json:"method,omitempty"`
	Handler http.HandlerFunc `json:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a new request router
func newReqRouter(host string) *reqrouter {
	r := new(reqrouter)
	r.host = host
	r.prefix = make(map[string]*reqs, defaultCap)
	return r
}

// Return a new set of requests
func newReqs(host, prefix string) *reqs {
	r := new(reqs)
	r.host = host
	r.prefix = prefix
	r.handlers = make([]*reqhandler, 0, defaultCap)
	return r
}

///////////////////////////////////////////////////////////////////////////////
// STRINIGFY

func (r reqhandler) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (router *reqrouter) AddHandler(ctx context.Context, prefix, path string, handler http.HandlerFunc, methods ...string) {
	// Add a path separator to the end of the prefix
	key := prefix
	if prefix != pathSep {
		key = prefix + pathSep
	}
	// Make a new set of requests associated with the prefix
	if _, exists := router.prefix[key]; !exists {
		router.prefix[key] = newReqs(router.host, prefix)
	}
	router.prefix[key].AddHandler(ctx, path, handler, methods...)
}

func (router *reqrouter) AddHandlerRe(ctx context.Context, prefix string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) {
	// Add a path separator to the end of the prefix
	key := prefix
	if prefix != pathSep {
		key = prefix + pathSep
	}

	// Make a new set of requests associated with the prefix
	if _, exists := router.prefix[key]; !exists {
		router.prefix[key] = newReqs(router.host, prefix)
	}
	router.prefix[key].AddHandlerRe(ctx, path, handler, methods...)
}

func (router *reqs) AddHandler(ctx context.Context, path string, handler http.HandlerFunc, methods ...string) {
	// Add any middleware to the handler, in reverse order
	middleware := Middleware(ctx)
	slices.Reverse(middleware)
	for _, middleware := range middleware {
		handler = middleware.Wrap(ctx, handler)
	}

	// Add the handler to the list
	router.handlers = append(router.handlers, &reqhandler{
		Label:   provider.Label(ctx),
		Host:    router.host,
		Prefix:  router.prefix,
		Path:    path,
		Handler: handler,
		Method:  methods,
	})
}

func (router *reqs) AddHandlerRe(ctx context.Context, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) {
	// Add any middleware to the handler, in reverse order
	middleware := Middleware(ctx)
	slices.Reverse(middleware)
	for _, middleware := range middleware {
		handler = middleware.Wrap(ctx, handler)
	}

	// Add the handler to the list
	router.handlers = append(router.handlers, &reqhandler{
		Label:   provider.Label(ctx),
		Host:    router.host,
		Prefix:  router.prefix,
		Re:      path,
		Handler: handler,
		Method:  methods,
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns a set of requests for a path
func (router *reqrouter) matchHandlers(path string) ([]*reqhandler, string) {
	// We assume that the path starts with a "/"
	for key, value := range router.prefix {
		// Match the prefix as a path
		if strings.HasPrefix(path, key) {
			if value.prefix != pathSep {
				// Remove the prefix from the path, before matching
				path = strings.TrimPrefix(path, value.prefix)
			}
			return value.matchHandlers(path), ""
		}
		// Match the prefix without a path, and redirect
		if path == value.prefix {
			return nil, value.prefix + pathSep
		}
	}

	// No match
	return nil, ""
}

// Return a list of handlers that match the path
func (router *reqs) matchHandlers(path string) []*reqhandler {
	// Assume the path starts with a "/" and the prefix has been removed
	result := make([]*reqhandler, 0, defaultCap)
	for _, handler := range router.handlers {
		if params := matchHandlerRe(handler, path); params != nil {
			result = append(result, handler)
		} else if matchHandlerPath(handler, path) {
			result = append(result, handler)
		}
	}
	return result
}

func matchHandlerRe(handler *reqhandler, path string) []string {
	if handler.Re == nil {
		return nil
	} else if params := handler.Re.FindStringSubmatch(path); params != nil {
		return params[1:]
	} else {
		return nil
	}
}

func matchHandlerPath(handler *reqhandler, path string) bool {
	if handler.Path == "" {
		return false
	}
	return strings.HasPrefix(path, handler.Path)
}

func matchMethod(handler *reqhandler, method string) bool {
	// Any method is allowed
	if len(handler.Method) == 0 {
		return true
	}
	// Specific methods are allowed
	return slices.Contains(handler.Method, method)
}
