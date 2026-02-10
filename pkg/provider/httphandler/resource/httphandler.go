package resource

import (
	"context"
	"net/http"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Resource describes an HTTP handler resource type. Each resource type is
// created via [NewResource] with a unique name, a hard-coded path, the
// handler function, and an optional OpenAPI spec. The handler is passive:
// its [Apply] simply stores the configuration. The router pulls the handler
// via the [httprouter.HandlerProvider] interface during its own Apply.
type Resource struct {
	Middleware bool   `name:"middleware" help:"Whether the router middleware chain wraps this handler" default:"true"`
	Endpoint   string `name:"endpoint" readonly:"" help:"Full URL endpoint for this handler (set by the router after Apply)"`
	name       string
	fn         http.HandlerFunc
	path       string
	spec       *openapi.PathItem
}

// ResourceInstance is a live instance of an HTTP handler resource.
type ResourceInstance struct {
	provider.ResourceInstance[Resource]
	fn         http.HandlerFunc
	path       string
	spec       *openapi.PathItem
	middleware bool
	router     schema.ResourceInstance // set via OnStateChange observer
}

var _ schema.Resource = Resource{}
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ httprouter.HandlerProvider = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewResource creates a handler resource type with the given unique name,
// path pattern (relative to the router prefix, e.g. "resource/{id}"),
// handler function, and optional OpenAPI path-item spec.
func NewResource(name, path string, fn http.HandlerFunc, spec *openapi.PathItem) Resource {
	return Resource{name: name, fn: fn, path: path, spec: spec}
}

func (r Resource) New() (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r),
		fn:               r.fn,
		path:             r.path,
		spec:             r.spec,
		middleware:       true,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (r Resource) Name() string {
	return r.name
}

func (r Resource) Schema() []schema.Attribute {
	return schema.AttributesOf(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// Validate decodes the incoming state, resolves references, and returns
// the validated *Resource configuration for use by Plan and Apply.
func (r *ResourceInstance) Validate(ctx context.Context, state schema.State, resolve schema.Resolver) (any, error) {
	return r.ResourceInstance.Validate(ctx, state, resolve)
}

// HandlerPath returns the route path relative to the router prefix.
func (r *ResourceInstance) HandlerPath() string {
	return r.path
}

// HandlerFunc returns the captured HTTP handler function.
func (r *ResourceInstance) HandlerFunc() http.HandlerFunc {
	return r.fn
}

// HandlerMiddleware reports whether middleware should wrap this handler.
func (r *ResourceInstance) HandlerMiddleware() bool {
	return r.middleware
}

// HandlerSpec returns the OpenAPI path-item for this handler, or nil.
func (r *ResourceInstance) HandlerSpec() *openapi.PathItem {
	return r.spec
}

// OnStateChange is called by the observer system when an instance
// that references this handler has its state changed. If the source
// is an httprouter, the reference is stored so [Read] can compute
// the endpoint dynamically.
func (r *ResourceInstance) OnStateChange(source schema.ResourceInstance) {
	if source.Resource().Name() == "httprouter" {
		r.router = source
	}
}

// Read returns the live state of the handler, computing the endpoint
// dynamically from the router's current state.
func (r *ResourceInstance) Read(ctx context.Context) (schema.State, error) {
	state, err := r.ResourceInstance.Read(ctx)
	if err != nil || state == nil {
		return state, err
	}
	if r.router != nil {
		if routerState, err := r.router.Read(ctx); err == nil && routerState != nil {
			if endpoint, ok := routerState["endpoint"].(string); ok {
				state["endpoint"] = endpoint + "/" + r.path
			}
		}
	}
	return state, nil
}

// Apply stores the configuration. The actual route registration is
// performed by the router during its own Apply.
func (r *ResourceInstance) Apply(_ context.Context, v any) error {
	c, ok := v.(*Resource)
	if !ok {
		return httpresponse.ErrInternalError.With("apply: unexpected config type")
	}
	r.middleware = c.Middleware
	r.SetStateAndNotify(c, r)
	return nil
}

// Destroy is a no-op.
func (r *ResourceInstance) Destroy(_ context.Context) error {
	return nil
}
