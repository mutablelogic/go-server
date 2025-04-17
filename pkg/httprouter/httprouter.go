package httprouter

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type router struct {
	*http.ServeMux
	prefix     string
	origin     string
	middleware []server.HTTPMiddleware
}

var _ server.Task = (*router)(nil)
var _ server.HTTPRouter = (*router)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ctx context.Context, prefix, origin string, middleware ...string) (*router, error) {
	router := new(router)
	router.ServeMux = http.NewServeMux()
	router.prefix = types.NormalisePath(prefix)
	router.origin = origin

	// Get middleware
	// TODO: Only store references to middleware here not the middleware itself
	for _, label := range middleware {
		middleware := ref.Provider(ctx).Task(ctx, label)
		if middleware == nil {
			return nil, httpresponse.ErrInternalError.Withf("%q is nil", label)
		} else if middleware_, ok := middleware.(server.HTTPMiddleware); !ok {
			return nil, httpresponse.ErrInternalError.Withf("%q is not HTTPMiddleware", label)
		} else {
			router.middleware = append(router.middleware, middleware_)
		}
	}

	// Return success
	return router, nil
}

////////////////////////////////////////////////////////////////////////////////
// TASK

func (r *router) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// HTTP ROUTER

// Register a function to handle a URL path
func (r *router) HandleFunc(ctx context.Context, prefix string, fn http.HandlerFunc) {
	// Wrap the function with middleware
	for _, middleware := range r.middleware {
		fn = middleware.HandleFunc(fn)
	}

	// Apply middleware, set context
	ref.Log(ctx).Debug(ctx, "Register route: ", types.JoinPath(r.prefix, prefix))
	r.ServeMux.HandleFunc(types.JoinPath(r.prefix, prefix), func(w http.ResponseWriter, r *http.Request) {
		fn(w, r.WithContext(ref.WithLog(ctx, ref.Log(ctx))))
	})
}

// Return the origin for CORS
func (r *router) Origin() string {
	return r.origin
}
