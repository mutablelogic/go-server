package resource

import (
	"context"
	"net/http"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Resource describes the "otel" middleware resource type. It is constructed
// via [NewResource] which captures the middleware function so that every
// instance created by [Resource.New] inherits it.
type Resource struct {
	fn httprouter.HTTPMiddlewareFunc
}

// ResourceInstance is a live instance of the otel middleware resource.
type ResourceInstance struct {
	provider.ResourceInstance[Resource]
	fn httprouter.HTTPMiddlewareFunc
}

var _ schema.Resource = Resource{}
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ httprouter.MiddlewareProvider = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewResource creates an otel middleware resource type that wraps every
// request with the given middleware function. The function is typically
// [otel.HTTPHandlerFunc(tracer)].
func NewResource(fn func(http.HandlerFunc) http.HandlerFunc) Resource {
	return Resource{fn: httprouter.HTTPMiddlewareFunc(fn)}
}

func (r Resource) New() (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r),
		fn:               r.fn,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (Resource) Name() string {
	return "otel"
}

func (Resource) Schema() []schema.Attribute {
	return schema.AttributesOf(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// Validate decodes the incoming state, resolves references, and returns
// the validated *Resource configuration for use by Plan and Apply.
func (r *ResourceInstance) Validate(ctx context.Context, state schema.State, resolve schema.Resolver) (any, error) {
	return r.ResourceInstance.Validate(ctx, state, resolve)
}

// MiddlewareFunc returns the captured middleware function. The router
// calls this during its own Apply to build the middleware chain.
func (r *ResourceInstance) MiddlewareFunc() httprouter.HTTPMiddlewareFunc {
	return r.fn
}

// Apply stores the configuration. The actual middleware attachment is
// performed by the router during its own Apply.
func (r *ResourceInstance) Apply(_ context.Context, v any) error {
	c, ok := v.(*Resource)
	if !ok {
		return httpresponse.ErrInternalError.With("apply: unexpected config type")
	}
	r.SetStateAndNotify(c, r)
	return nil
}

// Destroy is a no-op.
func (r *ResourceInstance) Destroy(_ context.Context) error {
	return nil
}
