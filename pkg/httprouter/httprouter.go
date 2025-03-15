package httprouter

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	provider "github.com/mutablelogic/go-server/pkg/provider"
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
	for _, name := range middleware {
		middleware, ok := provider.Provider(ctx).Task(ctx, name).(server.HTTPMiddleware)
		if !ok || middleware == nil {
			return nil, httpresponse.ErrInternalError.Withf("Invalid middleware %q", name)
		}
		router.middleware = append(router.middleware, middleware)
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
	provider.Log(ctx).Print(ctx, "Register route: ", types.JoinPath(r.prefix, prefix))
	r.ServeMux.HandleFunc(types.JoinPath(r.prefix, prefix), func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
		/* TODO fn(w, r.WithContext(
			provider.WithLog(
				provider.WithName(
					r.Context(), provider.Name(ctx),
				), provider.Log(ctx),
			),
		))*/
	})
}

// Return the origin for CORS
func (r *router) Origin() string {
	return r.origin
}
