package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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
	host := canonicalHost(r.Host)
	fmt.Printf("TODO: Implement AddHandler host=%q\n", host)
	httpresponse.Error(w, http.StatusNotFound, "Not found")
}

// Run the router until the context is cancelled
func (router *router) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (router *router) AddHandler(ctx context.Context, hostpath string, handler http.HandlerFunc, methods ...string) error {
	parts := strings.SplitN(hostpath, pathSep, 2)
	if len(parts) != 2 {
		return ErrBadParameter.Withf("AddHandler: %q", hostpath)
	}
	key := canonicalHost(parts[0])
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter()
	}
	return router.host[key].AddHandler(ctx, canonicalPrefix(ctx), pathSep+parts[1], handler, methods...)
}

func (router *router) AddHandlerRe(ctx context.Context, host string, path *regexp.Regexp, handler http.HandlerFunc, methods ...string) error {
	key := canonicalHost(host)
	if _, exists := router.host[key]; !exists {
		router.host[key] = newReqRouter()
	}
	return router.host[key].AddHandlerRe(ctx, canonicalPrefix(ctx), path, handler, methods...)
}

// Match handlers for a given method, host and path
func (router *router)Match(host, method, path string) (http.HandlerFunc, []string, int) {
	var results []*reqhandler

	// Check for host and path
	host = canonicalHost(host)
	for key, r := range router.host {
		if key != "" && !strings.HasSuffix(host, key) {
			continue
		}
		results = append(results, r.matchHandlers(path)...)
	}
	
	// Bail out if no results
	if len(results) == 0 {
		return nil, nil, http.StatusNotFound
	}

	// Check for methods
	for _, r := range results {
		if result.matchMethod(method) {



	// Return results
	return results
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
	return hostSep + strings.Trim(host, hostSep)
}
