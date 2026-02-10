package provider

import (
	"context"
	"fmt"
	"sync/atomic"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ResourceInstance provides common scaffolding for
// [schema.ResourceInstance] implementations.  The type parameter C is
// the concrete configuration struct that implements [schema.Resource].
//
// Embedding ResourceInstance[C] in a concrete type automatically
// satisfies the [Name], [Resource], [Plan], and [References] methods
// of the [schema.ResourceInstance] interface.  Concrete types must
// still implement [Validate], [Apply], and [Destroy].
type ResourceInstance[C schema.Resource] struct {
	name     string
	resource C
	state    *C
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var counter atomic.Int64

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewResourceInstance returns a ResourceInstance with an auto-generated
// instance name derived from the given resource type and a global
// monotonic counter. The resource value is stored so that [Resource]
// returns the actual type (including unexported fields such as the name
// on handler resources).
func NewResourceInstance[C schema.Resource](resource C) ResourceInstance[C] {
	return ResourceInstance[C]{
		name:     fmt.Sprintf("%s-%02d", resource.Name(), counter.Add(1)),
		resource: resource,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Name satisfies [schema.ResourceInstance].
func (b *ResourceInstance[C]) Name() string {
	return b.name
}

// SetName overrides the auto-generated name. This is used by
// [Manager.RegisterReadonlyInstance] to assign a deterministic label.
func (b *ResourceInstance[C]) SetName(name string) {
	b.name = name
}

// Resource satisfies [schema.ResourceInstance].  It returns the resource
// type that created this instance.
func (b *ResourceInstance[C]) Resource() schema.Resource {
	return b.resource
}

// State returns the last-applied configuration, or nil if the
// resource has not yet been applied.
func (b *ResourceInstance[C]) State() *C {
	return b.state
}

// SetState stores the applied configuration. Call this at the end
// of a successful [Apply].
func (b *ResourceInstance[C]) SetState(c *C) {
	b.state = c
}

// Validate decodes incoming [schema.State] into a *C, resolves
// references via resolve, and checks required/type constraints.
// Concrete Validate methods should call this first, then add
// resource-specific checks.
func (b *ResourceInstance[C]) Validate(_ context.Context, state schema.State, resolve schema.Resolver) (*C, error) {
	var desired C
	if err := state.Decode(&desired, resolve); err != nil {
		return nil, httpresponse.ErrBadRequest.Withf("invalid state: %v", err)
	}
	if err := schema.ValidateRefs(desired); err != nil {
		return nil, httpresponse.ErrBadRequest.With(err.Error())
	}
	if err := schema.ValidateRequired(desired); err != nil {
		return nil, httpresponse.ErrBadRequest.With(err.Error())
	}
	return &desired, nil
}

// Plan satisfies [schema.ResourceInstance].  It computes the diff between
// the validated configuration v (which must be *C) and the instance's
// current applied state.
func (b *ResourceInstance[C]) Plan(_ context.Context, v any) (schema.Plan, error) {
	desired, ok := v.(*C)
	if !ok {
		return schema.Plan{}, httpresponse.ErrInternalError.With("plan: unexpected config type")
	}
	newState := schema.StateOf(desired)

	// No current config means this is a new resource
	if b.state == nil {
		changes := make([]schema.Change, 0, len(newState))
		for field, val := range newState {
			changes = append(changes, schema.Change{
				Field: field,
				New:   val,
			})
		}
		return schema.Plan{Action: schema.ActionCreate, Changes: changes}, nil
	}

	// Compare each desired field against the current state
	oldState := schema.StateOf(b.state)
	var changes []schema.Change
	for field, newVal := range newState {
		oldVal := oldState[field]
		if fmt.Sprint(oldVal) != fmt.Sprint(newVal) {
			changes = append(changes, schema.Change{
				Field: field,
				Old:   oldVal,
				New:   newVal,
			})
		}
	}

	if len(changes) == 0 {
		return schema.Plan{Action: schema.ActionNoop}, nil
	}
	return schema.Plan{Action: schema.ActionUpdate, Changes: changes}, nil
}

// Read satisfies [schema.ResourceInstance].  It returns the live state
// of the instance by reflecting over the current applied configuration.
// Concrete types may override this to include computed fields.
func (b *ResourceInstance[C]) Read(_ context.Context) (schema.State, error) {
	if b.state == nil {
		return nil, nil
	}
	return schema.StateOf(b.state), nil
}

// References satisfies [schema.ResourceInstance].  It returns the labels
// of other resources this instance depends on.
func (b *ResourceInstance[C]) References() []string {
	if b.state == nil {
		return nil
	}
	return schema.ReferencesOf(*b.state)
}
