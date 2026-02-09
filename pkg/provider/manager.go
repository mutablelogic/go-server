package provider

import (
	"context"
	"errors"
	"strings"
	"sync"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	"github.com/mutablelogic/go-server/pkg/types"
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
		if err := inst.instance.Destroy(ctx); err != nil {
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
	}

	instanceName := resourceinstance.Name()
	if instanceName == "" || resourceinstance == nil {
		return nil, ErrBadRequest.With("resource instance is invalid")
	} else if _, exists := m.instances[instanceName]; exists {
		return nil, ErrConflict.Withf("resource instance %q already exists", instanceName)
	}

	// Set the instance
	m.instances[instanceName] = instance{
		instance: resourceinstance,
		state:    nil,
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

// ListResources returns metadata for every registered resource type
// (optionally filtered by type name) with their instances.
func (m *Manager) ListResources(_ context.Context, req schema.ListResourcesRequest) (*schema.ListResourcesResponse, error) {
	m.RLock()
	defer m.RUnlock()

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
			instances = append(instances, schema.InstanceMeta{
				Name:       inst.instance.Name(),
				Resource:   inst.instance.Resource().Name(),
				State:      inst.state,
				References: inst.instance.References(),
			})
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

	// Create and register the instance (newInstance stores it in m.instances)
	inst, err := m.newInstance(req.Resource)
	if err != nil {
		return nil, err
	}

	return &schema.CreateResourceInstanceResponse{
		Instance: schema.InstanceMeta{
			Name:       inst.Name(),
			Resource:   inst.Resource().Name(),
			References: inst.References(),
		},
	}, nil
}

// GetResourceInstance returns the metadata for a single named instance.
func (m *Manager) GetResourceInstance(_ context.Context, name string) (*schema.GetResourceInstanceResponse, error) {
	m.RLock()
	defer m.RUnlock()

	inst, exists := m.instances[name]
	if !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", name)
	}

	return &schema.GetResourceInstanceResponse{
		Instance: schema.InstanceMeta{
			Name:       inst.instance.Name(),
			Resource:   inst.instance.Resource().Name(),
			State:      inst.state,
			References: inst.instance.References(),
		},
	}, nil
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

	// Get the instance by name
	inst, exists := m.instances[name]
	if !exists {
		return nil, ErrNotFound.Withf("resource instance %q not found", name)
	}

	// Validate before planning/applying
	if err := inst.instance.Validate(ctx, req.Attributes); err != nil {
		return nil, err
	}

	// Compute the plan
	plan, err := inst.instance.Plan(ctx, req.Attributes)
	if err != nil {
		return nil, err
	}

	// Optionally apply the plan
	if req.Apply {
		newState, err := inst.instance.Apply(ctx, req.Attributes)
		if err != nil {
			return nil, err
		}
		m.instances[name] = instance{
			inst.instance,
			newState,
		}
		inst = m.instances[name]
	}

	return &schema.UpdateResourceInstanceResponse{
		Instance: schema.InstanceMeta{
			Name:       inst.instance.Name(),
			Resource:   inst.instance.Resource().Name(),
			State:      inst.state,
			References: inst.instance.References(),
		},
		Plan: plan,
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
	if err := inst.instance.Destroy(ctx); err != nil {
		return nil, err
	}

	// Remove from the manager
	delete(m.instances, req.Name)

	return &schema.DestroyResourceInstanceResponse{
		Name: req.Name,
	}, nil
}
