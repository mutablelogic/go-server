package resource

import (
	"log/slog"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Resource describes the "logger" middleware resource type. It is constructed
// via [NewResource] which captures the slog.Logger so that every instance
// created by [Resource.New] inherits it.
type Resource struct {
	fn httprouter.HTTPMiddlewareFunc
}

// ResourceInstance is a live instance of the logger middleware resource.
type ResourceInstance struct {
	provider.ResourceInstance[Resource]
	fn httprouter.HTTPMiddlewareFunc
}

var _ schema.Resource = Resource{}
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ server.HTTPMiddleware = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewResource creates a logger middleware resource type that wraps every
// HTTP request with structured request logging using the given logger.
// If log is nil, slog.Default() is used.
func NewResource(log *slog.Logger) Resource {
	return Resource{fn: httprouter.HTTPMiddlewareFunc(logger.NewMiddleware(log))}
}

func (r Resource) New(name string) (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r, name),
		fn:               r.fn,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (Resource) Name() string {
	return "logger"
}

func (Resource) Schema() []schema.Attribute {
	return schema.AttributesOf(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// WrapFunc wraps next with the captured middleware function.
// The router calls this during its own Apply to build the middleware chain.
func (r *ResourceInstance) WrapFunc(next http.HandlerFunc) http.HandlerFunc {
	return r.fn(next)
}
