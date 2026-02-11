package resource

import (
	"context"
	"net/http"
	"sort"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Resource struct {
	Prefix     string                    `name:"prefix" default:"/" help:"URL path prefix for all routes"`
	Origin     string                    `name:"origin" default:"" help:"Trusted origin for cross-origin requests (CSRF protection). Empty string means same-origin only, '*' allows all origins, or specify a scheme://host[:port]"`
	Title      string                    `name:"title" required:"" help:"OpenAPI spec title"`
	Version    string                    `name:"version" required:"" help:"OpenAPI spec version"`
	Endpoints  []string                  `name:"endpoints" readonly:"" help:"Base URLs of the router (one per attached server)"`
	NotFound   bool                      `name:"notfound" default:"true" help:"Register a default JSON 404 handler"`
	OpenAPI    bool                      `name:"openapi" default:"true" help:"Serve OpenAPI spec at {prefix}/openapi.json"`
	Middleware []schema.ResourceInstance `name:"middleware" help:"Ordered middleware instances to attach to the router"`
	Handlers   []schema.ResourceInstance `name:"handlers" help:"Handler instances to register on the router"`
}

type ResourceInstance struct {
	provider.ResourceInstance[Resource]
	router    *httprouter.Router
	mu        sync.RWMutex
	endpoints map[string]openapi.Server // server instance name -> OpenAPI server entry
}

var _ schema.Resource = (*Resource)(nil)
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ http.Handler = (*ResourceInstance)(nil)
var _ server.HTTPRouter = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	OpenAPIPath  = "/openapi.json"
	ResourceType = "httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r Resource) New() (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r),
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (Resource) Name() string {
	return ResourceType
}

func (Resource) Schema() []schema.Attribute {
	return schema.AttributesOf(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// Validate decodes the incoming state, resolves references, and returns
// the validated *Resource configuration for use by Plan and Apply.
func (r *ResourceInstance) Validate(ctx context.Context, state schema.State, resolve schema.Resolver) (any, error) {
	// Perform common validation and decoding using the embedded ResourceInstance
	desired, err := r.ResourceInstance.Validate(ctx, state, resolve)
	if err != nil {
		return nil, err
	}

	// Normalize prefix
	desired.Prefix = types.NormalisePath(desired.Prefix)

	// Return the validated configuration for use by Plan and Apply
	return desired, nil
}

// Apply materialises the resource using the validated configuration.
// It creates the router, attaches middleware in order, registers
// handlers, and adds default endpoints (404 and OpenAPI).
func (r *ResourceInstance) Apply(ctx context.Context, v any) error {
	c, err := r.ValidateConfig(v)
	if err != nil {
		return err
	}

	// Create a new router (no middleware passed to constructor — they come
	// from the middleware references below)
	router, err := httprouter.NewRouter(ctx, c.Prefix, c.Origin, c.Title, c.Version)
	if err != nil {
		return err
	}

	// Attach middleware in the order specified by the middleware attribute.
	// Each referenced instance must implement [server.HTTPMiddleware].
	for i, mw := range c.Middleware {
		mp, ok := mw.(server.HTTPMiddleware)
		if !ok {
			return httpresponse.ErrBadRequest.Withf("apply: middleware[%d] (%s) does not implement HTTPMiddleware", i, mw.Name())
		}
		router.AddMiddleware(mp.WrapFunc)
	}

	// Register handlers. Each referenced instance must implement
	// [server.HTTPHandler] or [server.HTTPFileServer]. Duplicate paths are rejected.
	seen := make(map[string]string, len(c.Handlers)) // path -> handler name
	for i, h := range c.Handlers {
		var path string
		switch hp := h.(type) {
		case server.HTTPFileServer:
			path = hp.HandlerPath()
			if prev, dup := seen[path]; dup {
				return httpresponse.ErrBadRequest.Withf("apply: duplicate handler path %q (handlers %s and %s)", path, prev, h.Name())
			}
			seen[path] = h.Name()
			router.RegisterFS(path, hp.HandlerFS(), true, hp.Spec())
		case server.HTTPHandler:
			path = hp.HandlerPath()
			if prev, dup := seen[path]; dup {
				return httpresponse.ErrBadRequest.Withf("apply: duplicate handler path %q (handlers %s and %s)", path, prev, h.Name())
			}
			seen[path] = h.Name()
			router.RegisterFunc(path, hp.HandlerFunc(), true, hp.Spec())
		default:
			return httpresponse.ErrBadRequest.Withf("apply: handlers[%d] (%s) does not implement HTTPHandler or HTTPFileServer", i, h.Name())
		}
	}

	// Register default endpoints
	if c.NotFound {
		router.RegisterNotFound("/", true)
	}
	if c.OpenAPI {
		router.RegisterOpenAPI(OpenAPIPath, true)
	}

	// Store the router and the applied config, notify observers
	r.router = router
	r.SetStateAndNotify(c, r)

	return nil
}

// Spec returns the OpenAPI specification for this router, or nil if
// the router has not been initialised yet.
func (r *ResourceInstance) Spec() *openapi.Spec {
	if r.router == nil {
		return nil
	}
	return r.router.Spec()
}

// OnStateChange is called by the observer system when an instance
// that references this router has its state changed. If the source
// implements [server.HTTPServer], its spec is stored and the OpenAPI
// spec's servers list is updated.
func (r *ResourceInstance) OnStateChange(source schema.ResourceInstance) {
	srv, ok := source.(server.HTTPServer)
	if !ok {
		return
	}
	spec := srv.Spec()
	if spec == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.endpoints == nil {
		r.endpoints = make(map[string]openapi.Server)
	}
	r.endpoints[source.Name()] = *spec
	r.syncServers()
}

// OnStateRemove is called by the observer system when an instance
// that references this router is being destroyed. If the source
// implements [server.HTTPServer], its entry is removed.
func (r *ResourceInstance) OnStateRemove(source schema.ResourceInstance) {
	if _, ok := source.(server.HTTPServer); ok {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.endpoints, source.Name())
		r.syncServers()
	}
}

// syncServers rebuilds the OpenAPI spec's servers list from the
// current endpoints map. Must be called with r.mu held.
func (r *ResourceInstance) syncServers() {
	if r.router == nil {
		return
	}
	c := r.State()
	if c == nil {
		return
	}
	servers := make([]openapi.Server, 0, len(r.endpoints))
	for _, srv := range r.endpoints {
		servers = append(servers, openapi.Server{URL: srv.URL + c.Prefix, Description: srv.Description})
	}
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].URL < servers[j].URL
	})
	r.router.Spec().SetServers(servers)
}

// Read returns the live state of the router, including the endpoints
// collected from attached servers.
func (r *ResourceInstance) Read(_ context.Context) (schema.State, error) {
	state, err := r.ResourceInstance.Read(context.Background())
	if err != nil || state == nil {
		return state, err
	}
	c := r.State()
	if c == nil {
		return state, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.endpoints) > 0 {
		endpoints := make([]string, 0, len(r.endpoints))
		for _, srv := range r.endpoints {
			endpoints = append(endpoints, srv.URL+c.Prefix)
		}
		sort.Strings(endpoints)
		state["endpoints"] = endpoints
	}
	return state, nil
}

// ServeHTTP delegates to the underlying router. This allows the
// resource instance to be used as an [http.Handler] by the httpserver.
func (r *ResourceInstance) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.router != nil {
		r.router.ServeHTTP(w, req)
	} else {
		http.Error(w, "router not initialised", http.StatusServiceUnavailable)
	}
}

// Destroy tears down the resource and releases its backing
// infrastructure. It returns an error if the resource cannot be
// cleanly removed.
func (r *ResourceInstance) Destroy(_ context.Context) error {
	r.router = nil
	return nil
}
