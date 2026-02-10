package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	ErrBadRequest = httpresponse.ErrBadRequest
	ErrNotFound   = httpresponse.ErrNotFound
	ErrConflict   = httpresponse.ErrConflict
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	name         string
	description  string
	version      string
	resources    map[string]schema.Resource
	instances    map[string]instance
	sync.RWMutex // Guard for instances
}

type instance struct {
	instance schema.ResourceInstance
	readOnly bool // manager-level override (e.g. RegisterReadonlyInstance)
}

var _ schema.Provider = (*Manager)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(name, description, version string) (*Manager, error) {
	return &Manager{
		name:        name,
		description: description,
		version:     version,
		resources:   make(map[string]schema.Resource),
		instances:   make(map[string]instance),
	}, nil
}

// Close destroys all instances and removes them from the manager,
// respecting dependency order (dependents are destroyed before their
// dependencies). It returns all errors encountered but continues
// destroying remaining instances.
func (m *Manager) Close(ctx context.Context) error {
	m.Lock()
	defer m.Unlock()

	// Build the set of all instance names
	all := make(map[string]bool, len(m.instances))
	for name := range m.instances {
		all[name] = true
	}

	var result error
	for _, name := range m.safeDestroyOrder(all) {
		inst, exists := m.instances[name]
		if !exists {
			continue
		}
		m.notifyRemovals(inst.instance)
		m.unwireObservers(inst.instance)
		if err := inst.instance.Destroy(ctx); err != nil {
			result = errors.Join(result, err)
		}
		delete(m.instances, name)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Name returns the unique name for the provider
func (m *Manager) Name() string {
	return m.name
}

// Description returns a human-readable summary of the provider.
func (m *Manager) Description() string {
	return m.description
}

// Version returns the semver of the provider.
func (m *Manager) Version() string {
	return m.version
}

// Resources returns the set of resources this provider can manage
func (m *Manager) Resources() []schema.Resource {
	m.RLock()
	defer m.RUnlock()
	resources := make([]schema.Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, r)
	}
	return resources
}

// New creates a resource instance from the given configuration.
// The returned resource is not yet applied; call [ResourceInstance.Apply] to
// materialise it.
func (m *Manager) New(name string) (schema.ResourceInstance, error) {
	m.Lock()
	defer m.Unlock()
	return m.newInstance(name)
}

// newInstance is the lock-free core of [New]. The caller must hold m.Lock().
func (m *Manager) newInstance(name string) (schema.ResourceInstance, error) {
	// Look up the resource type
	resource, exists := m.resources[name]
	if !exists {
		return nil, ErrBadRequest.Withf("resource %q is not registered", name)
	}

	// Create a new resource instance
	resourceinstance, err := resource.New()
	if err != nil {
		return nil, fmt.Errorf("resource %q: %w", name, err)
	}
	if resourceinstance == nil {
		return nil, ErrBadRequest.With("resource instance is nil")
	}

	instanceName := resourceinstance.Name()
	if instanceName == "" {
		return nil, ErrBadRequest.With("resource instance name is empty")
	} else if _, exists := m.instances[instanceName]; exists {
		return nil, ErrConflict.Withf("resource instance %q already exists", instanceName)
	}

	// Set the instance
	m.instances[instanceName] = instance{
		instance: resourceinstance,
	}

	// Return the instance
	return resourceinstance, nil
}

// RegisterResource registers a resource type with the provider. It returns
// an error if a resource with the same name is already registered.
func (m *Manager) RegisterResource(r schema.Resource) error {
	m.Lock()
	defer m.Unlock()
	if r == nil {
		return ErrBadRequest.With("resource is nil")
	} else if name := r.Name(); name == "" {
		return ErrBadRequest.With("resource name is empty")
	} else if _, exists := m.resources[name]; exists {
		return ErrConflict.Withf("resource %q is already registered", name)
	} else {
		m.resources[name] = r
	}

	// Return success
	return nil
}

// RegisterReadonlyInstance creates a new instance of the given resource
// type with a deterministic label, validates and applies the given
// state, and stores the result as a read-only (data) instance. The
// instance name is "{resource}-{label}". If the resource type is not
// yet registered it is registered automatically. It returns the created
// [schema.ResourceInstance] so the caller can type-assert to the
// concrete object (e.g. *httpserver.Server).
func (m *Manager) RegisterReadonlyInstance(ctx context.Context, resource schema.Resource, label string, state schema.State) (schema.ResourceInstance, error) {
	m.Lock()
	defer m.Unlock()

	// Auto-register the resource type if not already registered
	name := resource.Name()
	if _, exists := m.resources[name]; !exists {
		m.resources[name] = resource
	}

	// Create a new instance
	inst, err := resource.New()
	if err != nil {
		return nil, fmt.Errorf("resource %q: %w", name, err)
	}

	// Override the auto-generated name with {resource}-{label}
	instanceName := name + "-" + label
	if nameable, ok := inst.(interface{ SetName(string) }); ok {
		nameable.SetName(instanceName)
	}
	if inst.Name() != instanceName {
		return nil, ErrBadRequest.Withf("instance does not support SetName")
	}
	if _, exists := m.instances[instanceName]; exists {
		return nil, ErrConflict.Withf("instance %q already exists", instanceName)
	}

	// Store temporarily so the resolver can find it during Validate
	m.instances[instanceName] = instance{instance: inst}

	// Validate the state
	config, err := inst.Validate(ctx, state, m.resolver())
	if err != nil {
		delete(m.instances, instanceName)
		return nil, fmt.Errorf("instance %q: validate: %w", instanceName, err)
	}

	// Apply the state
	if err := inst.Apply(ctx, config); err != nil {
		delete(m.instances, instanceName)
		return nil, fmt.Errorf("instance %q: apply: %w", instanceName, err)
	}

	// Store with read-only flag
	m.instances[instanceName] = instance{
		instance: inst,
		readOnly: true,
	}

	// Wire observers so referenced instances are notified on state changes
	m.wireAndNotify(inst)

	return inst, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE QUERIES

// ListResources returns metadata for every registered resource type
// (optionally filtered by type name) with their instances.
func (m *Manager) ListResources(ctx context.Context, req schema.ListResourcesRequest) (*schema.ListResourcesResponse, error) {
	m.RLock()
	defer m.RUnlock()

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Get the filter for the resource type
	filter := strings.TrimSpace(types.Value(req.Type))
	if filter != "" {
		if _, exists := m.resources[filter]; !exists {
			return nil, ErrBadRequest.Withf("resource type %q is not registered", filter)
		}
	}
	resources := make([]schema.ResourceMeta, 0, len(m.resources))
	for _, r := range m.resources {
		if filter != "" && r.Name() != filter {
			continue
		}

		// Collect instances for this resource type
		instances := make([]schema.InstanceMeta, 0)
		for _, inst := range m.instances {
			if inst.instance.Resource().Name() != r.Name() {
				continue
			}
			instances = append(instances, m.instanceMeta(ctx, inst))
		}

		resources = append(resources, schema.ResourceMeta{
			Name:       r.Name(),
			Attributes: r.Schema(),
			Instances:  instances,
		})
	}
	return &schema.ListResourcesResponse{
		Provider:    m.name,
		Description: m.description,
		Version:     m.version,
		Resources:   resources,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - INSTANCE LIFECYCLE

// CreateResourceInstance creates a new instance from the named resource type
// and stores it in the manager. The instance is not yet validated or applied;
// call [UpdateResourceInstance] to plan or apply it.
func (m *Manager) CreateResourceInstance(ctx context.Context, req schema.CreateResourceInstanceRequest) (*schema.CreateResourceInstanceResponse, error) {
	m.Lock()
	defer m.Unlock()

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Create and register the instance (newInstance stores it in m.instances)
	inst, err := m.newInstance(req.Resource)
	if err != nil {
		return nil, err
	}

	return &schema.CreateResourceInstanceResponse{
		Instance: m.instanceMeta(ctx, m.instances[inst.Name()]),
	}, nil
}

// GetResourceInstance returns the metadata for a single named instance.
func (m *Manager) GetResourceInstance(ctx context.Context, name string) (*schema.GetResourceInstanceResponse, error) {
	m.RLock()
	defer m.RUnlock()

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check instance, return its metadata
	if inst, exists := m.instances[name]; !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", name)
	} else {
		return &schema.GetResourceInstanceResponse{
			Instance: m.instanceMeta(ctx, inst),
		}, nil
	}
}

// UpdateResourceInstance validates and plans the named instance. When
// req.Apply is true the plan is also applied and the resulting state stored.
func (m *Manager) UpdateResourceInstance(ctx context.Context, name string, req schema.UpdateResourceInstanceRequest) (*schema.UpdateResourceInstanceResponse, error) {
	// Apply needs a write lock; plan is read-only
	if req.Apply {
		m.Lock()
		defer m.Unlock()
	} else {
		m.RLock()
		defer m.RUnlock()
	}

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Get the instance by name
	inst, exists := m.instances[name]
	if !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", name)
	}

	// Reject updates to read-only (data) instances
	if inst.readOnly {
		return nil, ErrConflict.Withf("cannot update %q: instance is read-only", name)
	}

	// Validate before planning/applying — returns the decoded config
	config, err := inst.instance.Validate(ctx, req.Attributes, m.resolver())
	if err != nil {
		return nil, fmt.Errorf("instance %q: validate: %w", name, err)
	}

	// Reject circular dependencies
	if err := m.checkCycles(name, inst.instance.Resource().Schema(), req.Attributes); err != nil {
		return nil, err
	}

	// Compute the plan
	plan, err := inst.instance.Plan(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("instance %q: plan: %w", name, err)
	}

	// Optionally apply the plan
	if req.Apply {
		// Remove stale observers and notify old references before
		// applying the new config, since Apply will overwrite the
		// instance's references.
		m.notifyRemovals(inst.instance)
		m.unwireObservers(inst.instance)

		if err := inst.instance.Apply(ctx, config); err != nil {
			// Re-wire old observers on failure so state stays consistent
			m.wireAndNotify(inst.instance)
			return nil, fmt.Errorf("instance %q: apply: %w", name, err)
		}
		m.instances[name] = instance{
			instance: inst.instance,
			readOnly: inst.readOnly,
		}
		inst = m.instances[name]
		m.wireAndNotify(inst.instance)
	}

	return &schema.UpdateResourceInstanceResponse{
		Instance: m.instanceMeta(ctx, inst),
		Plan:     plan,
	}, nil
}

// DestroyResourceInstance tears down the named instance and removes it
// from the manager. When req.Cascade is true, all instances that
// (transitively) depend on the target are destroyed first, in
// topological order.
func (m *Manager) DestroyResourceInstance(ctx context.Context, req schema.DestroyResourceInstanceRequest) (*schema.DestroyResourceInstanceResponse, error) {
	m.Lock()
	defer m.Unlock()

	// Check context
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check the target exists
	if _, exists := m.instances[req.Name]; !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", req.Name)
	}

	// Reject destruction of read-only (data) instances
	inst := m.instances[req.Name]
	if inst.readOnly {
		return nil, ErrConflict.Withf("cannot destroy %q: instance is read-only", req.Name)
	}

	// Build the ordered list of instances to destroy
	var order []string
	if req.Cascade {
		subgraph := make(map[string]bool)
		m.collectDependents(req.Name, subgraph)
		order = m.safeDestroyOrder(subgraph)
	} else {
		// Without cascade, refuse if anything depends on this instance
		if err := m.checkDependents(req.Name); err != nil {
			return nil, err
		}
		order = []string{req.Name}
	}

	// Destroy each instance in order, collecting metadata
	destroyed := make([]schema.InstanceMeta, 0, len(order))
	for _, name := range order {
		inst, exists := m.instances[name]
		if !exists {
			continue
		}

		// Read-only instances cannot be destroyed via cascade
		if inst.readOnly {
			return nil, ErrConflict.Withf("cannot destroy %q: instance is read-only", name)
		}

		// Capture metadata before destruction
		meta := m.instanceMeta(ctx, inst)

		m.notifyRemovals(inst.instance)
		m.unwireObservers(inst.instance)
		if err := inst.instance.Destroy(ctx); err != nil {
			return nil, err
		}
		delete(m.instances, name)
		destroyed = append(destroyed, meta)
	}

	return &schema.DestroyResourceInstanceResponse{
		Instances: destroyed,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// instanceMeta builds an [schema.InstanceMeta] from a live instance.
// It calls Read() on the instance to get the current state.
// The caller must hold at least m.RLock().
func (m *Manager) instanceMeta(ctx context.Context, inst instance) schema.InstanceMeta {
	var state schema.State
	if s, err := inst.instance.Read(ctx); err == nil {
		state = s
	}
	return schema.InstanceMeta{
		Name:       inst.instance.Name(),
		Resource:   inst.instance.Resource().Name(),
		ReadOnly:   inst.readOnly,
		State:      state,
		References: inst.instance.References(),
	}
}

// directDependents returns the names of all live instances that
// directly reference the given instance name. The caller must hold
// at least m.RLock().
func (m *Manager) directDependents(name string) []string {
	var deps []string
	for instName, inst := range m.instances {
		if instName == name {
			continue
		}
		for _, ref := range inst.instance.References() {
			if ref == name {
				deps = append(deps, instName)
				break
			}
		}
	}
	return deps
}

// collectDependents recursively adds name and every instance that
// (transitively) depends on it to the set. The caller must hold at
// least m.RLock().
func (m *Manager) collectDependents(name string, set map[string]bool) {
	if set[name] {
		return
	}
	set[name] = true
	for _, dep := range m.directDependents(name) {
		m.collectDependents(dep, set)
	}
}

// safeDestroyOrder returns the given set of instance names in
// safe-to-destroy order: instances that nothing else (within the set)
// depends on come first, working inward to the foundations. Any
// instances involved in cycles are appended at the end so they are
// still destroyed. The caller must hold at least m.RLock().
func (m *Manager) safeDestroyOrder(names map[string]bool) []string {
	// Build adjacency within the set.
	// dependsOn[X] = refs X makes to other set members.
	// inDegree[X]  = how many set members depend on X.
	dependsOn := make(map[string][]string, len(names))
	inDegree := make(map[string]int, len(names))
	for name := range names {
		inDegree[name] = 0
	}
	for name := range names {
		inst := m.instances[name]
		for _, ref := range inst.instance.References() {
			if names[ref] {
				dependsOn[name] = append(dependsOn[name], ref)
				inDegree[ref]++
			}
		}
	}

	// Kahn's algorithm — nodes that nothing depends on are safe to
	// destroy first, working inward.
	queue := make([]string, 0, len(names))
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	order := make([]string, 0, len(names))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		order = append(order, name)
		for _, ref := range dependsOn[name] {
			inDegree[ref]--
			if inDegree[ref] == 0 {
				queue = append(queue, ref)
			}
		}
	}

	// Any remaining instances form a cycle — append them so they
	// still get destroyed.
	for name := range names {
		if inDegree[name] > 0 {
			order = append(order, name)
		}
	}

	return order
}

// resolver returns a [schema.Resolver] that looks up instances by name.
// The caller must hold at least m.RLock().
func (m *Manager) resolver() schema.Resolver {
	return func(name string) schema.ResourceInstance {
		if inst, ok := m.instances[name]; ok {
			return inst.instance
		}
		return nil
	}
}

// checkCycles returns an error if the prospective references in state would
// create a circular dependency through instanceName. It extracts reference
// names from the state using the resource schema, then walks the existing
// dependency graph via DFS to see if any path leads back to instanceName.
// The caller must hold at least m.RLock().
func (m *Manager) checkCycles(instanceName string, attrs []schema.Attribute, state schema.State) error {
	newRefs := schema.ReferencesFromState(attrs, state)
	if len(newRefs) == 0 {
		return nil
	}

	visited := make(map[string]bool, len(m.instances))
	var walk func(string) bool
	walk = func(current string) bool {
		if current == instanceName {
			return true
		}
		if visited[current] {
			return false
		}
		visited[current] = true
		if inst, ok := m.instances[current]; ok {
			for _, ref := range inst.instance.References() {
				if walk(ref) {
					return true
				}
			}
		}
		return false
	}

	for _, ref := range newRefs {
		if walk(ref) {
			return ErrBadRequest.Withf("circular dependency: %q references %q which leads back to %q", instanceName, ref, instanceName)
		}
	}
	return nil
}

// wireAndNotify registers observers on the applied instance so that
// its referenced (dependency) instances are notified when the applied
// instance's state changes in the future. It also fires an initial
// notification for each reference, since SetState already ran during
// Apply before the observers were wired.
// The caller must hold m.Lock().
func (m *Manager) wireAndNotify(inst schema.ResourceInstance) {
	obs, ok := inst.(Observable)
	if !ok {
		return
	}
	for _, refName := range inst.References() {
		dep, exists := m.instances[refName]
		if !exists {
			continue
		}
		depInst := dep.instance
		obs.AddObserver(refName, func(source schema.ResourceInstance) {
			if h, ok := depInst.(interface {
				OnStateChange(schema.ResourceInstance)
			}); ok {
				h.OnStateChange(source)
			}
		})
		// Fire initial notification
		if h, ok := depInst.(interface {
			OnStateChange(schema.ResourceInstance)
		}); ok {
			h.OnStateChange(inst)
		}
	}
}

// notifyRemovals tells each referenced (dependency) instance that the
// given instance is being removed. This is the inverse of the initial
// OnStateChange notification in wireAndNotify.
// The caller must hold m.Lock().
func (m *Manager) notifyRemovals(inst schema.ResourceInstance) {
	for _, refName := range inst.References() {
		dep, exists := m.instances[refName]
		if !exists {
			continue
		}
		if h, ok := dep.instance.(interface {
			OnStateRemove(schema.ResourceInstance)
		}); ok {
			h.OnStateRemove(inst)
		}
	}
}

// unwireObservers removes observers that reference the given instance.
// For each depender (an instance that references this one), the observer
// with the destroyed instance's name is removed.
// The caller must hold m.Lock().
func (m *Manager) unwireObservers(inst schema.ResourceInstance) {
	name := inst.Name()
	for _, dependerName := range m.directDependents(name) {
		depender, ok := m.instances[dependerName]
		if !ok {
			continue
		}
		if obs, ok := depender.instance.(Observable); ok {
			obs.RemoveObserver(name)
		}
	}
}

// checkDependents returns an error if any live instance depends on the
// given instance name. The caller must hold at least m.RLock().
func (m *Manager) checkDependents(name string) error {
	if deps := m.directDependents(name); len(deps) > 0 {
		return ErrConflict.Withf("cannot destroy %q: instance %q depends on it", name, deps[0])
	}
	return nil
}
