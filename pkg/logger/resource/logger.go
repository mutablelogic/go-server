package resource

import (
	"context"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Resource describes a "logger" for logging and middleware
type Resource struct {
	Debug  bool   `name:"debug" help:"Enable debug logging" default:"false"`
	Format string `name:"format" help:"Log format" enum:"text,term,json" default:"text"`
}

// ResourceInstance is a live instance of the log middleware resource.
type ResourceInstance struct {
	provider.ResourceInstance[Resource]
	*logger.Logger
}

var _ schema.Resource = Resource{}
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ server.HTTPMiddleware = (*ResourceInstance)(nil)
var _ server.Logger = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	resourceType = "logger"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r Resource) New(name string) (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance(r, name),
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

// Apply creates or reconfigures the logger from the validated configuration.
func (r *ResourceInstance) Apply(ctx context.Context, v any) error {
	return r.ApplyConfig(ctx, v, func(ctx context.Context, c *Resource) error {
		// Create or reconfigure the logger
		if r.Logger == nil {
			r.Logger = logger.New(os.Stderr, logger.FormatFromString(c.Format), c.Debug)
		} else {
			r.Logger.SetDebug(c.Debug)
		}
		return nil
	})
}

// Destroy releases the logger.
func (r *ResourceInstance) Destroy(_ context.Context) error {
	r.Logger = nil
	return nil
}
