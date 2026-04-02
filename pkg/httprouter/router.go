// Package httprouter provides an HTTP request router with built-in cross-origin
// request forgery (CSRF) protection, middleware support, and OpenAPI spec generation.
// Routes are registered under a common path prefix and may optionally pass through
// a middleware chain before reaching the handler.
package httprouter

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi_ops "github.com/mutablelogic/go-server/pkg/openapi"
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
	security   map[string]SecurityScheme
}

var _ http.Handler = (*Router)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewRouter creates a new router with the given prefix, origin, title, version
// and middleware. The prefix is normalised to a path. The origin controls
// cross-origin request handling:
//   - Empty string: same-origin requests only (CSRF enabled, no CORS)
//   - "*": allow all cross-origin requests (CORS enabled, CSRF bypassed)
//   - A specific origin in the form "scheme://host[:port]" (CORS and CSRF
//     enabled, with the origin added as a trusted CSRF origin)
//
// The title and version are used to create the OpenAPI spec for the router.
func NewRouter(ctx context.Context, mux *http.ServeMux, prefix, origin, title, version string, middleware ...HTTPMiddlewareFunc) (*Router, error) {
	if mux == nil {
		return nil, httpresponse.ErrBadRequest.With("mux is nil")
	}

	router := new(Router)
	router.mux = mux
	router.prefix = types.NormalisePath(prefix)
	router.origin = origin
	router.middleware = middlewareFuncs(middleware)
	router.security = make(map[string]SecurityScheme)

	// Create a new OpenAPI spec
	router.spec = openapi.NewSpec(title, version)

	// Build the handler chain depending on the origin policy.
	// When origin is "*" CSRF is bypassed entirely because there is no
	// meaningful CSRF protection when all origins are trusted.
	switch {
	case origin == "*":
		// All origins trusted – CORS only, no CSRF
		router.handler = Cors(origin)(router.mux.ServeHTTP)
	case origin != "":
		// Specific origin – CSRF with trusted origin, wrapped by CORS
		crf := http.NewCrossOriginProtection()
		if err := crf.AddTrustedOrigin(origin); err != nil {
			return nil, err
		}
		router.handler = Cors(origin)(crf.Handler(router.mux).ServeHTTP)
	default:
		// Empty origin – same-origin only, CSRF with no trusted origins
		router.handler = http.NewCrossOriginProtection().Handler(router.mux)
	}

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

// Prefix returns the normalised path prefix that all relative routes are
// registered under.
func (r *Router) Prefix() string {
	return r.prefix
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

// resolvePath returns path unchanged when it is absolute (starts with "/"),
// otherwise it joins it with the router's prefix.
func (r *Router) resolvePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return types.JoinPath(r.prefix, path)
}

// safeHandle registers a handler with the router's ServeMux, recovering from
// panics caused by duplicate patterns. Returns an error instead of panicking.
func (r *Router) safeHandle(pattern string, handler http.HandlerFunc) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = httpresponse.ErrConflict.Withf("%v", v)
		}
	}()
	r.mux.HandleFunc(pattern, handler)
	return nil
}

// RegisterCatchAll registers a structured-404 handler at the literal "/" path
// in the underlying ServeMux. Because http.ServeMux treats "/" as a catch-all
// that matches any request not handled by a more-specific pattern, this ensures
// unmatched requests return a JSON 404 instead of Go's plain-text response.
func (r *Router) RegisterCatchAll(path string, middleware bool) error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = httpresponse.Error(w, httpresponse.ErrNotFound, req.RequestURI)
	})
	if middleware {
		handler = r.middleware.Wrap(handler)
	}
	if err := r.safeHandle(types.NormalisePath(path), handler); err != nil {
		if errors.Is(err, httpresponse.ErrConflict) {
			return nil
		}
		return err
	}
	return nil
}

// RegisterOpenAPI registers a handler at path that serves the router's
// [openapi.Spec] as JSON on GET requests and returns 405 Method Not Allowed
// for all other HTTP methods. If path is relative the router prefix is
// prepended; if path is absolute it is used as-is. When middleware is true
// the handler is wrapped by the router's middleware chain.
func (r *Router) RegisterOpenAPI(path string, middleware bool) error {
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
	return r.safeHandle(r.resolvePath(path), handler)
}

// RegisterFS registers a file server at path that serves static assets from
// the given [fs.FS]. If path is relative (does not start with "/") the router
// prefix is prepended; if path is absolute it is used as-is. The combined
// prefix is stripped from incoming requests before the file lookup. A trailing
// slash is ensured so that [http.ServeMux] treats it as a subtree pattern,
// matching all sub-paths.
//
// When spec is non-nil the corresponding [openapi.PathItem] is added to the
// router's OpenAPI specification under the resolved path. When middleware is
// true the handler is wrapped by the router's middleware chain.
func (r *Router) RegisterFS(path string, fs fs.FS, middleware bool, spec *openapi.PathItem) error {
	prefix := r.resolvePath(path)
	if prefix != "/" {
		prefix += "/"
	}
	if spec != nil {
		r.spec.AddPath(r.resolvePath(path), spec)
	}
	handler := http.StripPrefix(prefix, http.FileServer(http.FS(fs))).ServeHTTP
	if middleware {
		handler = r.middleware.Wrap(handler)
	}
	return r.safeHandle(prefix, handler)
}

// RegisterPath registers a [PathItem] handler at path. If path is relative
// the router prefix is prepended. Any security schemes referenced by the
// path item's OpenAPI operations must already be registered on the router;
// matching handlers are wrapped with those security schemes before the
// router's middleware chain is applied.
func (r *Router) RegisterPath(path string, params *jsonschema.Schema, pathitem httprequest.PathItem) error {
	// Resolve the path with the router prefix
	path = r.resolvePath(path)

	// OpenAPI spec is optional, but if provided, add the path to the spec
	// and wrap per-method handlers with their security requirements
	spec := pathitem.Spec(path, params)
	if spec != nil {
		var registerErr error

		// Look at security requirements for each operation and apply any
		// corresponding middleware to the per-method handler
		openapi_ops.Operations(spec, func(method string, op *openapi.Operation) {
			if registerErr != nil {
				return
			}
			for _, requirement := range op.Security {
				for name, scopes := range requirement {
					scheme, ok := r.security[name]
					if !ok {
						registerErr = httpresponse.ErrNotImplemented.Withf("security scheme %q not registered for %s %s", name, method, path)
						return
					}
					pathitem.WrapHandler(method, func(next http.HandlerFunc) http.HandlerFunc {
						return scheme.Wrap(next, scopes)
					})
				}
			}
		})
		if registerErr != nil {
			return registerErr
		}
		r.spec.AddPath(path, spec)
	}

	// Get handler or fall back to method-not-allowed
	handler := pathitem.Handler()
	if handler == nil {
		return httpresponse.ErrNotImplemented.Withf("path item %q has no handlers", path)
	}

	// Register the handler
	return r.safeHandle(path, r.middleware.Wrap(handler))
}
