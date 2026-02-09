package schema

import (
	"net/url"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// LIST RESOURCES

// ListResourcesRequest filters the set of resource types to return.
type ListResourcesRequest struct{}

// ListResourcesResponse contains the resource types the provider can manage.
type ListResourcesResponse struct {
	Provider    string         `json:"provider"`
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version,omitempty"`
	Resources   []ResourceMeta `json:"resources"`
}

// ResourceMeta describes a resource type without exposing the full interface.
type ResourceMeta struct {
	Name       string      `json:"name"`
	Attributes []Attribute `json:"attributes"`
}

///////////////////////////////////////////////////////////////////////////////
// LIST RESOURCE INSTANCES

// ListResourceInstancesRequest filters the set of instances to return.
// An empty Resource field returns all instances.
type ListResourceInstancesRequest struct {
	Resource string `json:"resource,omitempty"` // filter by resource type name
}

// ListResourceInstancesResponse contains the live resource instances.
type ListResourceInstancesResponse struct {
	Instances []InstanceMeta `json:"instances"`
}

// InstanceMeta describes a resource instance without exposing the full interface.
type InstanceMeta struct {
	Name       string   `json:"name"`
	Resource   string   `json:"resource"`
	State      State    `json:"state,omitempty"`
	References []string `json:"references,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// CREATE RESOURCE INSTANCE

// CreateResourceInstanceRequest asks the manager to instantiate a resource.
type CreateResourceInstanceRequest struct {
	Resource string `json:"resource" arg:"" help:"Resource type name (e.g. \"httpserver\")"` // resource type name (e.g. "httpserver")
}

// CreateResourceInstanceResponse contains the newly created instance metadata.
type CreateResourceInstanceResponse struct {
	Instance InstanceMeta `json:"instance"`
}

///////////////////////////////////////////////////////////////////////////////
// PLAN RESOURCE INSTANCE

// PlanResourceInstanceRequest asks the manager to compute a plan for an instance.
type PlanResourceInstanceRequest struct {
	Name string `json:"name"` // instance name
}

// PlanResourceInstanceResponse contains the computed plan.
type PlanResourceInstanceResponse struct {
	Instance InstanceMeta `json:"instance"`
	Plan     Plan         `json:"plan"`
}

///////////////////////////////////////////////////////////////////////////////
// APPLY RESOURCE INSTANCE

// ApplyResourceInstanceRequest asks the manager to apply (create/update) an instance.
type ApplyResourceInstanceRequest struct {
	Name string `json:"name"` // instance name
}

// ApplyResourceInstanceResponse contains the new state after apply.
type ApplyResourceInstanceResponse struct {
	Instance InstanceMeta `json:"instance"`
}

///////////////////////////////////////////////////////////////////////////////
// DESTROY RESOURCE INSTANCE

// DestroyResourceInstanceRequest asks the manager to tear down an instance.
type DestroyResourceInstanceRequest struct {
	Name string `json:"name"` // instance name
}

// DestroyResourceInstanceResponse confirms destruction.
type DestroyResourceInstanceResponse struct {
	Name string `json:"name"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (req ListResourcesResponse) String() string {
	return types.Stringify(req)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r ListResourcesRequest) Query() url.Values {
	return url.Values{}
}
