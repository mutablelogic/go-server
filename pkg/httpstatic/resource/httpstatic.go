package resource

import (
	"context"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpstatic "github.com/mutablelogic/go-server/pkg/httpstatic"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Resource describes a static file server resource type. The Dir field
// specifies the directory on disk to serve. The Path field determines the
// URL path relative to the router prefix. Summary and Description are
// surfaced via the OpenAPI specification.
type Resource struct {
	Path        string   `name:"path" required:"" help:"URL path relative to the router prefix"`
	Dir         string   `name:"dir" required:"" help:"Directory on disk to serve"`
	Summary     string   `name:"summary" default:"" help:"Short summary for the OpenAPI spec"`
	Description string   `name:"description" default:"" help:"Description for the OpenAPI spec"`
	Endpoints   []string `name:"endpoints" readonly:"" help:"Full URL endpoints for this handler"`
}

// ResourceInstance is a live instance of a static file server resource.
type ResourceInstance struct {
	provider.ResourceInstance[Resource]
	static  server.HTTPFileServer
	mu      sync.RWMutex
	routers map[string]schema.ResourceInstance // keyed by router instance name
}

var _ schema.Resource = (*Resource)(nil)
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ server.HTTPFileServer = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	resourceType = "httpstatic"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r Resource) New(name string) (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r, name),
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (Resource) Name() string {
	return resourceType
}

func (Resource) Schema() []schema.Attribute {
	return schema.AttributesOf(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// Validate decodes the incoming state, resolves references, and returns
// the validated *Resource configuration for use by Plan and Apply.
func (r *ResourceInstance) Validate(ctx context.Context, state schema.State, resolve schema.Resolver) (any, error) {
	v, err := r.ResourceInstance.Validate(ctx, state, resolve)
	if err != nil {
		return nil, err
	}
	desired := v.(*Resource)

	// Normalise the path
	desired.Path = types.NormalisePath(desired.Path)

	// Return the validated configuration
	return desired, nil
}

// OnStateChange is called by the observer system when an instance
// that references this handler has its state changed. If the source
// is an httprouter, the reference is stored so [Read] can compute
// the endpoints dynamically.
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
				path := r.HandlerPath()
				for _, ep := range eps {
					endpoints = append(endpoints, strings.TrimRight(ep, "/")+types.NormalisePath(path))
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

// Apply materialises the resource using the validated configuration.
// It creates an [httpstatic.Static] backed by [os.DirFS] for the
// configured directory.
func (r *ResourceInstance) Apply(ctx context.Context, v any) error {
	return r.ApplyConfig(ctx, v, func(ctx context.Context, c *Resource) error {
		// Create the static file server from the directory on disk
		static, err := httpstatic.New(c.Path, os.DirFS(c.Dir), c.Summary, c.Description)
		if err != nil {
			return err
		}
		r.static = static
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - HTTP HANDLER

// HandlerPath returns the route path relative to the router prefix.
func (r *ResourceInstance) HandlerPath() string {
	return r.static.HandlerPath()
}

// HandlerFS returns the filesystem to serve.
func (r *ResourceInstance) HandlerFS() fs.FS {
	return r.static.HandlerFS()
}

// HandlerSpec returns the OpenAPI path-item for this handler, or nil.
func (r *ResourceInstance) Spec() *openapi.PathItem {
	return r.static.Spec()
}
