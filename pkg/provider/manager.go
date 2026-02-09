package provider

import (
	"context"
	"errors"
	"sync"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
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
	state    schema.State
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

	var result error
	for _, name := range m.destroyOrder() {
		inst, exists := m.instances[name]
		if !exists {
			continue
		}
		if err := inst.instance.Destroy(ctx, inst.state); err != nil {
			result = errors.Join(result, err)
		}
		delete(m.instances, name)
	}
	return result
}

// destroyOrder returns instance names in reverse-dependency order: instances
// that depend on others come first so they are torn down before their
// dependencies. The caller must hold at least m.RLock().
func (m *Manager) destroyOrder() []string {
	// Build an adjacency set: for each instance, which other instances
	// depend on it (i.e. its "dependents").
	dependents := make(map[string][]string, len(m.instances))
	inDegree := make(map[string]int, len(m.instances))
	for name := range m.instances {
		inDegree[name] = 0
	}
	for name, inst := range m.instances {
		for _, ref := range inst.instance.References() {
			if _, exists := m.instances[ref]; exists {
				dependents[ref] = append(dependents[ref], name)
				inDegree[name]++
			}
		}
	}

	// Kahn's algorithm — start with instances that have no dependencies
	// (leaves), which means they are dependents of others and should be
	// destroyed first.
	queue := make([]string, 0, len(m.instances))
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	order := make([]string, 0, len(m.instances))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		order = append(order, name)
		for _, dep := range dependents[name] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// Any remaining instances form a cycle — append them anyway so they
	// still get destroyed.
	for name := range m.instances {
		if inDegree[name] > 0 {
			order = append(order, name)
		}
	}

	return order
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
		return nil, err
	} else if name := resourceinstance.Name(); name == "" || resourceinstance == nil {
		return nil, ErrBadRequest.With("resource instance is invalid")
	} else if _, exists := m.instances[name]; exists {
		return nil, ErrConflict.Withf("resource instance %q already exists", resourceinstance.Name())
	}

	// Set the instance
	m.instances[name] = instance{
		instance: resourceinstance,
	}

	// Return the instance
	return resourceinstance, nil
}

// RegisterResource registers a resource type with the provider. It returns
// an error if a resource with the same name is already registered.
func (m *Manager) RegisterResource(r schema.Resource) error {
	if r == nil {
		return ErrBadRequest.With("resource is nil")
	}
	name := r.Name()
	if name == "" {
		return ErrBadRequest.With("resource name is empty")
	}
	if _, exists := m.resources[name]; exists {
		return ErrConflict.Withf("resource %q is already registered", name)
	}
	m.resources[name] = r
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE QUERIES

// ListResources returns metadata for every registered resource type.
func (m *Manager) ListResources(_ context.Context, _ schema.ListResourcesRequest) (*schema.ListResourcesResponse, error) {
	resources := make([]schema.ResourceMeta, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, schema.ResourceMeta{
			Name:       r.Name(),
			Attributes: r.Schema(),
		})
	}
	return &schema.ListResourcesResponse{
		Provider:    m.name,
		Description: m.description,
		Version:     m.version,
		Resources:   resources,
	}, nil
}

// ListResourceInstances returns metadata for live instances, optionally
// filtered by resource type.
func (m *Manager) ListResourceInstances(_ context.Context, req schema.ListResourceInstancesRequest) (*schema.ListResourceInstancesResponse, error) {
	m.RLock()
	defer m.RUnlock()

	instances := make([]schema.InstanceMeta, 0, len(m.instances))
	for _, inst := range m.instances {
		// Apply resource-type filter if specified
		if req.Resource != "" && inst.instance.Resource().Name() != req.Resource {
			continue
		}
		instances = append(instances, schema.InstanceMeta{
			Name:       inst.instance.Name(),
			Resource:   inst.instance.Resource().Name(),
			State:      inst.state,
			References: inst.instance.References(),
		})
	}
	return &schema.ListResourceInstancesResponse{Instances: instances}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - INSTANCE LIFECYCLE

// CreateResourceInstance creates a new instance from the named resource type,
// validates it, and stores it in the manager. The instance is not yet applied.
func (m *Manager) CreateResourceInstance(ctx context.Context, req schema.CreateResourceInstanceRequest) (*schema.CreateResourceInstanceResponse, error) {
	m.Lock()
	defer m.Unlock()

	// Create a new instance
	inst, err := m.newInstance(req.Resource)
	if err != nil {
		return nil, err
	} else if err := inst.Validate(ctx); err != nil {
		return nil, err
	}

	// Store the instance
	m.instances[inst.Name()] = instance{
		instance: inst,
	}

	return &schema.CreateResourceInstanceResponse{
		Instance: schema.InstanceMeta{
			Name:       inst.Name(),
			Resource:   inst.Resource().Name(),
			References: inst.References(),
		},
	}, nil
}

// PlanResourceInstance computes the changes needed to bring the named
// instance to its desired state without modifying anything.
func (m *Manager) PlanResourceInstance(ctx context.Context, req schema.PlanResourceInstanceRequest) (*schema.PlanResourceInstanceResponse, error) {
	m.RLock()
	defer m.RUnlock()

	// Get an instance by name
	inst, exists := m.instances[req.Name]
	if !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", req.Name)
	}

	// Plan the instance changes
	plan, err := inst.instance.Plan(ctx, inst.state)
	if err != nil {
		return nil, err
	}

	return &schema.PlanResourceInstanceResponse{
		Instance: schema.InstanceMeta{
			Name:       inst.instance.Name(),
			Resource:   inst.instance.Resource().Name(),
			State:      inst.state,
			References: inst.instance.References(),
		},
		Plan: plan,
	}, nil
}

// ApplyResourceInstance applies the named instance, creating or updating
// its backing infrastructure and recording the resulting state.
func (m *Manager) ApplyResourceInstance(ctx context.Context, req schema.ApplyResourceInstanceRequest) (*schema.ApplyResourceInstanceResponse, error) {
	m.Lock()
	defer m.Unlock()

	// Get an instance by name
	inst, exists := m.instances[req.Name]
	if !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", req.Name)
	}

	// Apply and capture the new state
	newState, err := inst.instance.Apply(ctx, inst.state)
	if err != nil {
		return nil, err
	}

	// Update the stored state
	inst.state = newState
	m.instances[req.Name] = inst

	return &schema.ApplyResourceInstanceResponse{
		Instance: schema.InstanceMeta{
			Name:       inst.instance.Name(),
			Resource:   inst.instance.Resource().Name(),
			State:      newState,
			References: inst.instance.References(),
		},
	}, nil
}

// DestroyResourceInstance tears down the named instance and removes it
// from the manager.
func (m *Manager) DestroyResourceInstance(ctx context.Context, req schema.DestroyResourceInstanceRequest) (*schema.DestroyResourceInstanceResponse, error) {
	m.Lock()
	defer m.Unlock()

	// Get the instance by name
	inst, exists := m.instances[req.Name]
	if !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", req.Name)
	}

	// Destroy the backing infrastructure
	if err := inst.instance.Destroy(ctx, inst.state); err != nil {
		return nil, err
	}

	// Remove from the manager
	delete(m.instances, req.Name)

	return &schema.DestroyResourceInstanceResponse{
		Name: req.Name,
	}, nil
}
