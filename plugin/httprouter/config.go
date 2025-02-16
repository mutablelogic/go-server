package httprouter

import (
	"context"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/httprouter"
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Prefix     string   `kong:"-"`
	Origin     string   `default:"*" help:"CORS origin"`
	Middleware []string `default:"" help:"Middleware to apply to all routes"`
}

var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	return httprouter.New(ctx, c.Prefix, c.Origin, c.Middleware...)
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return "httprouter"
}

func (c Config) Description() string {
	return "HTTP request router"
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
	r.ServeMux.HandleFunc(types.JoinPath(r.prefix, prefix), func(w http.ResponseWriter, r *http.Request) {
		fn(w, r.WithContext(
			provider.WithLog(
				provider.WithName(
					r.Context(), provider.Name(ctx),
				), provider.Log(ctx),
			),
		))
	})
}

// Return the origin for CORS
func (r *router) Origin() string {
	return r.origin
}
