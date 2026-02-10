// Package httprouter provides an HTTP request router with built-in cross-origin
// request forgery (CSRF) protection, middleware support, and OpenAPI spec generation.
// Routes are registered under a common path prefix and may optionally pass through
// a middleware chain before reaching the handler.
package httprouter

import (
	"context"
	"io/fs"
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Router is an HTTP request multiplexer that wraps [http.ServeMux] with
// cross-origin protection, an optional middleware chain and an [openapi.Spec]
// that is populated as routes are registered.
type Router struct {
	mux        *http.ServeMux
	prefix     string
	origin     string
	middleware middlewareFuncs
	handler    http.Handler
	spec       *openapi.Spec
}

var _ http.Handler = (*Router)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewRouter creates a new router with the given prefix, origin, title, version
// and middleware. The prefix is normalised to a path. The origin is used for
// cross-origin protection (CSRF) and can be:
//   - Empty string: same-origin requests only
//   - "*": allow all cross-origin requests
//   - A specific origin in the form "scheme://host[:port]"
//
// The title and version are used to create the OpenAPI spec for the router.
func NewRouter(ctx context.Context, prefix, origin, title, version string, middleware ...HTTPMiddlewareFunc) (*Router, error) {
	router := new(Router)
	router.mux = http.NewServeMux()
	router.prefix = types.NormalisePath(prefix)
	router.origin = origin
	router.middleware = middlewareFuncs(middleware)

	// Create a new OpenAPI spec
	router.spec = openapi.NewSpec(title, version)

	// Create a CSRF handler, and set the router's handler to the CSRF handler
	// Origin can be empty (same-origin only), "*" (allow all cross-origin)
	// or a specific origin in the form "scheme://host[:port]"
	crf := http.NewCrossOriginProtection()
	switch {
	case origin == "", origin == "*":
		// No trusted origin added
	default:
		if err := crf.AddTrustedOrigin(origin); err != nil {
			return nil, err
		}
	}
	router.handler = crf.Handler(router.mux)

	// Return success
	return router, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AddMiddleware appends fn to the router's middleware chain. Middleware is
// applied in order: the first added becomes the outermost wrapper and
// executes first when a request arrives.  This method must be called before
// any routes are registered, because routes capture the chain at
// registration time.
func (r *Router) AddMiddleware(fn HTTPMiddlewareFunc) {
	r.middleware = append(r.middleware, fn)
}

// Origin returns the trusted origin configured for cross-origin protection.
// It returns an empty string when only same-origin requests are allowed,
// "*" when all origins are trusted, or a specific "scheme://host[:port]" value.
func (r *Router) Origin() string {
	return r.origin
}

// Spec returns the OpenAPI 3.1 specification for this router. The spec is
// built incrementally as routes are registered via [router.RegisterFunc].
func (r *Router) Spec() *openapi.Spec {
	return r.spec
}

// ServeHTTP dispatches the request to the matching registered handler after
// applying cross-origin protection. It implements the [http.Handler] interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

// RegisterNotFound registers a catch-all handler at path that responds with
// 404 Not Found. It is typically registered at "/" so that any request that
// does not match a more specific route receives a structured error response.
// When middleware is true the handler is wrapped by the router's middleware chain.
func (r *Router) RegisterNotFound(path string, middleware bool) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		_ = httpresponse.Error(w, httpresponse.ErrNotFound, req.RequestURI)
	}
	if middleware {
		handler = r.middleware.Wrap(handler)
	}
	r.mux.HandleFunc(types.JoinPath(r.prefix, path), handler)
}

// RegisterOpenAPI registers a handler at path that serves the router's
// [openapi.Spec] as JSON on GET requests and returns 405 Method Not Allowed
// for all other HTTP methods. When middleware is true the handler is wrapped
// by the router's middleware chain.
func (r *Router) RegisterOpenAPI(path string, middleware bool) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(req), r.spec)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), req.Method)
		}
	}
	if middleware {
		handler = r.middleware.Wrap(handler)
	}
	r.mux.HandleFunc(types.JoinPath(r.prefix, path), handler)
}

// RegisterFS registers a file server at path that serves static assets from
// the given [fs.FS]. The router prefix is prepended to path and stripped from
// incoming requests before the file lookup. When middleware is true the
// handler is wrapped by the router's middleware chain.
func (r *Router) RegisterFS(path string, fs fs.FS, middleware bool) {
	handler := http.StripPrefix(types.JoinPath(r.prefix, path), http.FileServer(http.FS(fs))).ServeHTTP
	if middleware {
		handler = r.middleware.Wrap(handler)
	}
	r.mux.HandleFunc(types.JoinPath(r.prefix, path), handler)
}

// RegisterFunc registers handler at path. The path should not include an HTTP
// method; the handler itself is responsible for differentiating between methods.
// The router prefix is prepended to path before registration.
//
// When spec is non-nil the corresponding [openapi.PathItem] is added to the
// router's OpenAPI specification under the resolved path. When middleware is
// true the handler is wrapped by the router's middleware chain.
func (r *Router) RegisterFunc(path string, handler http.HandlerFunc, middleware bool, spec *openapi.PathItem) {
	// OpenAPI spec is optional, but if provided, add the path to the spec
	path = types.JoinPath(r.prefix, path)
	if spec != nil {
		r.spec.AddPath(path, spec)
	}

	// Optionally add middleware to the handler, and register the handler
	if middleware {
		handler = r.middleware.Wrap(handler)
	}

	// Register the handler
	r.mux.HandleFunc(path, handler)
}
