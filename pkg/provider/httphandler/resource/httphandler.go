package resource

import (
	"context"
	"net/http"
	"sort"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
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
// via the [server.HTTPHandler] interface during its own Apply.
type Resource struct {
	Middleware bool     `name:"middleware" help:"Whether the router middleware chain wraps this handler" default:"true"`
	Endpoints  []string `name:"endpoints" readonly:"" help:"Full URL endpoints for this handler"`
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
	mu         sync.RWMutex
	routers    map[string]schema.ResourceInstance // keyed by router instance name
}

var _ schema.Resource = Resource{}
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ server.HTTPHandler = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewResource creates a handler resource type with the given unique name,
// path pattern (relative to the router prefix, e.g. "resource/{id}"),
// handler function, and optional OpenAPI path-item spec.
func NewResource(name, path string, fn http.HandlerFunc, spec *openapi.PathItem) Resource {
	return Resource{name: name, fn: fn, path: path, spec: spec}
}

func (r Resource) New(name string) (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r, name),
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
func (r *ResourceInstance) Spec() *openapi.PathItem {
	return r.spec
}

// OnStateChange is called by the observer system when an instance
// that references this handler has its state changed. If the source
// is an httprouter, the reference is stored so [Read] can compute
// the endpoints dynamically. Multiple routers can reference the
// same handler, so they are stored in a map keyed by instance name.
func (r *ResourceInstance) OnStateChange(source schema.ResourceInstance) {
	if source.Resource().Name() == "httprouter" {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.routers == nil {
			r.routers = make(map[string]schema.ResourceInstance)
		}
		r.routers[source.Name()] = source
	}
}

// OnStateRemove is called by the observer system when an instance
// that references this handler is being destroyed. If the source
// is an httprouter, its reference is removed.
func (r *ResourceInstance) OnStateRemove(source schema.ResourceInstance) {
	if source.Resource().Name() == "httprouter" {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.routers, source.Name())
	}
}

// Read returns the live state of the handler, computing the endpoints
// dynamically from all attached routers' current state.
func (r *ResourceInstance) Read(ctx context.Context) (schema.State, error) {
	state, err := r.ResourceInstance.Read(ctx)
	if err != nil || state == nil {
		return state, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	var endpoints []string
	for _, rtr := range r.routers {
		if rtrState, err := rtr.Read(ctx); err == nil && rtrState != nil {
			if eps, ok := rtrState["endpoints"].([]string); ok {
				for _, ep := range eps {
					endpoints = append(endpoints, ep+"/"+r.path)
				}
			}
		}
	}
	sort.Strings(endpoints)
	if len(endpoints) > 0 {
		state["endpoints"] = endpoints
	}
	return state, nil
}

// Apply stores the configuration. The actual route registration is
// performed by the router during its own Apply.
func (r *ResourceInstance) Apply(ctx context.Context, v any) error {
	return r.ApplyConfig(ctx, v, func(ctx context.Context, c *Resource) error {
		r.middleware = c.Middleware
		return nil
	})
}
