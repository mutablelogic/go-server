package provider

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// OBSERVER TYPES

// ObserverFunc is called when the instance's state changes.
// The source parameter is the concrete [schema.ResourceInstance]
// whose state was updated.
type ObserverFunc func(source schema.ResourceInstance)

// Observable is optionally satisfied by resource instances that
// support state-change observer registration. All types that embed
// [ResourceInstance] automatically implement this interface.
type Observable interface {
	AddObserver(id string, fn ObserverFunc)
	RemoveObserver(id string)
}

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
	name      string
	resource  C
	state     atomic.Pointer[C]
	mu        sync.RWMutex
	observers map[string]ObserverFunc
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
	return b.state.Load()
}

// SetState stores the applied configuration without notifying
// observers. Prefer [SetStateAndNotify] in Apply methods.
func (b *ResourceInstance[C]) SetState(c *C) {
	b.state.Store(c)
}

// ValidateConfig asserts that v is a *C and returns it. Use this at
// the top of Apply methods to replace the duplicated type-assertion
// boilerplate.
func (b *ResourceInstance[C]) ValidateConfig(v any) (*C, error) {
	c, ok := v.(*C)
	if !ok {
		return nil, httpresponse.ErrInternalError.With("apply: unexpected config type")
	}
	return c, nil
}

// SetStateAndNotify stores the applied configuration and notifies
// all registered observers. The source parameter should be the
// outermost [schema.ResourceInstance] (typically the concrete type
// that embeds ResourceInstance). Call this at the end of a
// successful [Apply].
func (b *ResourceInstance[C]) SetStateAndNotify(c *C, source schema.ResourceInstance) {
	b.state.Store(c)
	b.NotifyObservers(source)
}

// AddObserver registers a callback that will be invoked when this
// instance's state changes via [SetStateAndNotify]. The id is
// typically the name of the observing instance; calling AddObserver
// with an existing id replaces the previous callback.
func (b *ResourceInstance[C]) AddObserver(id string, fn ObserverFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.observers == nil {
		b.observers = make(map[string]ObserverFunc)
	}
	b.observers[id] = fn
}

// RemoveObserver unregisters the observer with the given id.
func (b *ResourceInstance[C]) RemoveObserver(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.observers, id)
}

// NotifyObservers calls all registered observer callbacks with the
// given source instance. The observer map is copied under the lock
// so callbacks are free to call AddObserver/RemoveObserver without
// deadlocking.
func (b *ResourceInstance[C]) NotifyObservers(source schema.ResourceInstance) {
	b.mu.RLock()
	snapshot := make([]ObserverFunc, 0, len(b.observers))
	for _, fn := range b.observers {
		snapshot = append(snapshot, fn)
	}
	b.mu.RUnlock()
	for _, fn := range snapshot {
		fn(source)
	}
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
	current := b.state.Load()
	if current == nil {
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
	oldState := schema.StateOf(current)
	var changes []schema.Change
	for field, newVal := range newState {
		oldVal := oldState[field]
		if !reflect.DeepEqual(oldVal, newVal) {
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
	current := b.state.Load()
	if current == nil {
		return nil, nil
	}
	return schema.StateOf(current), nil
}

// References satisfies [schema.ResourceInstance].  It returns the labels
// of other resources this instance depends on.
func (b *ResourceInstance[C]) References() []string {
	current := b.state.Load()
	if current == nil {
		return nil
	}
	return schema.ReferencesOf(*current)
}
