package router

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	// Packages
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Represents a set of routes which are matched by a host/prefix combination
type reqrouter struct {
	// Maps prefixes to an array of handlers
	prefix map[string]*reqs
}

// Represents a set of handlers for a prefix
type reqs struct {
	prefix   string
	handlers []*reqhandler
}

// Represents a handler for a request
type reqhandler struct {
	Path    string           `json:"path,omitempty"`
	Re      *regexp.Regexp   `json:"re,omitempty"`
	Method  []string         `json:"method,omitempty"`
	Handler http.HandlerFunc `json:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a new request router
func newReqRouter() *reqrouter {
	r := new(reqrouter)
	r.prefix = make(map[string]*reqs, defaultCap)
	return r
}

// Return a new set of requests
func newReqs(prefix string) *reqs {
	r := new(reqs)
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

func (router *reqrouter) AddHandler(ctx context.Context, prefix, path string, handler http.HandlerFunc, methods ...string) error {
	// Add a path separator to the end of the prefix
	key := prefix
	if prefix != pathSep {
		key = prefix + pathSep
	}
	// Make a new set of requests associated with the prefix
	if _, exists := router.prefix[prefix]; !exists {
		router.prefix[key] = newReqs(prefix)
	}
	return router.prefix[key].AddHandler(ctx, path, handler, methods...)
}

func (router *reqrouter) AddHandlerRe(ctx context.Context, prefix string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	// Add a path separator to the end of the prefix
	key := prefix
	if prefix != pathSep {
		key = prefix + pathSep
	}
	// Make a new set of requests associated with the prefix
	if _, exists := router.prefix[prefix]; !exists {
		router.prefix[key] = newReqs(prefix)
	}
	return router.prefix[key].AddHandlerRe(ctx, path, handler, methods...)
}

func (router *reqs) AddHandler(ctx context.Context, path string, handler http.HandlerFunc, methods ...string) error {
	router.handlers = append(router.handlers, &reqhandler{
		Path:    path,
		Handler: handler,
		Method:  methods,
	})
	return nil
}

func (router *reqs) AddHandlerRe(ctx context.Context, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	router.handlers = append(router.handlers, &reqhandler{
		Re:      path,
		Handler: handler,
		Method:  methods,
	})
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns a set of requests for a path
func (router *reqrouter) matchHandlers(path string) []*reqhandler {
	// We assume that the path starts with a "/"
	for key, value := range router.prefix {
		if strings.HasPrefix(path, key) {
			if value.prefix != pathSep {
				// Remove the prefix from the path, before matching
				path = strings.TrimPrefix(path, value.prefix)
			}
			return value.matchHandlers(path)
		}
	}

	// No match
	return nil
}

// Return a list of handlers that match the path
func (router *reqs) matchHandlers(path string) []*reqhandler {
	// Assume the path starts with a "/" and the prefix has been removed
	result := make([]*reqhandler, 0, defaultCap)
	for _, handler := range router.handlers {
		if handler.Re != nil && matchHandlerRe(handler, path) {
			result = append(result, handler)
		} else if matchHandlerPath(handler, path) {
			result = append(result, handler)
		}
	}
	return result
}

func matchHandlerRe(handler *reqhandler, path string) bool {
	return handler.Re.MatchString(path)
}

func matchHandlerPath(handler *reqhandler, path string) bool {
	return handler.Path == path
}
