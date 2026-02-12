package schema

import "context"

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Provider is a named provider that registers one or more resource types.
type Provider interface {
	// Name returns the unique name for the provider.
	Name() string

	// Description returns a human-readable summary of the provider.
	Description() string

	// Resources returns the set of resources this provider can manage.
	Resources() []Resource
}

// Resource describes a kind of resource a provider can manage.
// It acts as a factory: it advertises the configuration schema
// and creates new instances.
type Resource interface {
	// Name returns a unique name for this resource type, used in
	// configuration and plan output.
	Name() string

	// Schema returns the attribute definitions for this resource type,
	// used for validation and documentation.
	Schema() []Attribute

	// New creates a named resource instance.
	// The returned instance is not yet applied; call
	// [ResourceInstance.Apply] to materialise it.
	New(name string) (ResourceInstance, error)
}

// ResourceInstance represents a single configured unit of infrastructure
// with a full create/read/update/delete lifecycle.
type ResourceInstance interface {
	// Name returns the logical name of this resource instance
	// (e.g. "main", "api-v1").
	Name() string

	// Resource returns the resource type that created this instance.
	Resource() Resource

	// Validate decodes the incoming [State] using the given [Resolver],
	// checks that the resulting configuration is complete and consistent,
	// and returns the typed configuration value (as any) for use by
	// [Plan] and [Apply].
	Validate(context.Context, State, Resolver) (any, error)

	// Plan computes the difference between the validated configuration
	// (returned by [Validate]) and the instance's current state,
	// returning a set of planned changes without modifying anything.
	Plan(context.Context, any) (Plan, error)

	// Apply materialises the resource using the validated configuration
	// (returned by [Validate]), creating or updating it to match the
	// desired state.
	Apply(context.Context, any) error

	// Destroy tears down the resource and releases its backing
	// infrastructure. It returns an error if the resource cannot be
	// cleanly removed.
	Destroy(context.Context) error

	// Read returns the current live state of the instance. This is
	// used to refresh the manager's cached state and to detect drift
	// from the desired configuration. Implementations may re-read
	// computed fields (e.g. endpoint URLs) from the running infra.
	Read(context.Context) (State, error)

	// References returns the labels of other resources this resource
	// depends on. The runtime must ensure those resources are applied
	// first and destroyed last.
	References() []string
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Action describes the kind of change planned for a resource.
type Action string

const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDestroy Action = "destroy"
	ActionNoop    Action = "noop"
)

// Plan describes the changes that [Resource.Apply] would make.
type Plan struct {
	// Action is the high-level operation (create, update, destroy, noop).
	Action Action

	// Changes lists the individual field-level diffs. It is empty when
	// Action is Noop or Destroy.
	Changes []Change
}

// Change describes a single field-level diff within a [Plan].
type Change struct {
	// Field is the attribute name.
	Field string

	// Old is the current value (nil on create).
	Old any

	// New is the desired value (nil on destroy).
	New any
}
