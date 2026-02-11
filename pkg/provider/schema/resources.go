package schema

import (
	"net/url"
	"strings"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// LIST RESOURCES

// ListResourcesRequest filters the set of resource types to return.
type ListResourcesRequest struct {
	Type *string `json:"type,omitempty" arg:"" optional:"" help:"Resource type name (e.g. \"httpserver\")"` // filter by type name
}

// ListResourcesResponse contains the resource types the provider can manage.
type ListResourcesResponse struct {
	Provider    string         `json:"provider"`
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version,omitempty"`
	Resources   []ResourceMeta `json:"resources"`
}

// ResourceMeta describes a resource type without exposing the full interface.
type ResourceMeta struct {
	Name       string         `json:"name"`
	Attributes []Attribute    `json:"attributes"`
	Instances  []InstanceMeta `json:"instances"`
}

// InstanceMeta describes a resource instance without exposing the full interface.
type InstanceMeta struct {
	Name       string   `json:"name"`
	Resource   string   `json:"resource"`
	ReadOnly   bool     `json:"readonly,omitempty"`
	State      State    `json:"state,omitempty"`
	References []string `json:"references,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// CREATE RESOURCE INSTANCE

// CreateResourceInstanceRequest asks the manager to instantiate a resource.
type CreateResourceInstanceRequest struct {
	Resource string `json:"resource" arg:"" help:"Resource type name (e.g. \"httpserver\")"`
	Label    string `json:"label" arg:"" help:"Instance label (e.g. \"main\")"`
}

// CreateResourceInstanceResponse contains the newly created instance metadata.
type CreateResourceInstanceResponse struct {
	Instance InstanceMeta `json:"instance"`
}

///////////////////////////////////////////////////////////////////////////////
// GET RESOURCE INSTANCE

// GetResourceInstanceResponse contains the metadata for a single instance.
type GetResourceInstanceResponse struct {
	Instance InstanceMeta `json:"instance"`
}

///////////////////////////////////////////////////////////////////////////////
// UPDATE RESOURCE INSTANCE (PLAN / APPLY)

// UpdateResourceInstanceRequest is the unified request for plan or apply.
// When Apply is false (the default), only a plan is computed. When true,
// the planned changes are applied. Attributes contains the desired attribute
// values keyed by attribute name.
type UpdateResourceInstanceRequest struct {
	Attributes State `json:"attributes,omitempty"` // desired attribute values
	Apply      bool  `json:"apply"`                // false = plan only, true = apply changes
}

// UpdateResourceInstanceResponse contains the instance metadata, the computed
// plan, and (when apply was requested) the resulting state.
type UpdateResourceInstanceResponse struct {
	Instance InstanceMeta `json:"instance"`
	Plan     Plan         `json:"plan"`
}

///////////////////////////////////////////////////////////////////////////////
// DESTROY RESOURCE INSTANCE

// DestroyResourceInstanceRequest asks the manager to tear down an instance.
type DestroyResourceInstanceRequest struct {
	Name    string `json:"name"`              // instance name
	Cascade bool   `json:"cascade,omitempty"` // also destroy dependents in topological order
}

// DestroyResourceInstanceResponse confirms destruction and returns the
// metadata of every destroyed instance (in destruction order).
type DestroyResourceInstanceResponse struct {
	Instances []InstanceMeta `json:"instances"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r ListResourcesResponse) String() string {
	return types.Stringify(r)
}

func (r CreateResourceInstanceResponse) String() string {
	return types.Stringify(r)
}

func (r GetResourceInstanceResponse) String() string {
	return types.Stringify(r)
}

func (r UpdateResourceInstanceResponse) String() string {
	return types.Stringify(r)
}

func (r DestroyResourceInstanceResponse) String() string {
	return types.Stringify(r)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r ListResourcesRequest) Query() url.Values {
	v := url.Values{}
	if t := strings.TrimSpace(types.Value(r.Type)); t != "" {
		v.Set("type", t)
	}
	return v
}
