package provider_test

import (
	"context"
	"testing"

	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/mutablelogic/go-server/pkg/provider/schema"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// MOCK TYPES

type mockResource struct {
	name string
}

func (m *mockResource) Name() string               { return m.name }
func (m *mockResource) Schema() []schema.Attribute { return schema.AttributesOf(nodeConfig{}) }
func (m *mockResource) New(name string) (schema.ResourceInstance, error) {
	return &mockInstance{
		name:     name,
		resource: m,
	}, nil
}

// nodeConfig is a resource config with an optional dependency on another instance.
type nodeConfig struct {
	Dep schema.ResourceInstance `name:"dep" type:"node"`
}

type mockInstance struct {
	name     string
	resource schema.Resource
	refs     []string
}

func (m *mockInstance) Name() string              { return m.name }
func (m *mockInstance) Resource() schema.Resource { return m.resource }
func (m *mockInstance) Validate(_ context.Context, _ schema.State, _ schema.Resolver) (any, error) {
	return nil, nil
}
func (m *mockInstance) Plan(_ context.Context, _ any) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockInstance) Apply(_ context.Context, _ any) error {
	return nil
}
func (m *mockInstance) Destroy(_ context.Context) error { return nil }
func (m *mockInstance) Read(_ context.Context) (schema.State, error) {
	return nil, nil
}
func (m *mockInstance) References() []string { return m.refs }

///////////////////////////////////////////////////////////////////////////////
// TESTS - CYCLE DETECTION

func Test_Manager_Cycle_001(t *testing.T) {
	assert := assert.New(t)

	// A -> B is fine (no cycle)
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	// Update A to depend on B — should succeed
	_, err = mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{"dep": b.Name()},
		Apply:      true,
	})
	assert.NoError(err)
}

func Test_Manager_Cycle_002(t *testing.T) {
	assert := assert.New(t)

	// A -> B -> A is a cycle
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	// A depends on B
	a.(*mockInstance).refs = []string{b.Name()}

	// Now try to make B depend on A — cycle
	_, err = mgr.UpdateResourceInstance(context.Background(), b.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{"dep": a.Name()},
		Apply:      true,
	})
	assert.Error(err)
	assert.Contains(err.Error(), "circular")
}

func Test_Manager_Cycle_003(t *testing.T) {
	assert := assert.New(t)

	// Self-reference: A -> A is a cycle
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)

	_, err = mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{"dep": a.Name()},
		Apply:      true,
	})
	assert.Error(err)
	assert.Contains(err.Error(), "circular")
}

func Test_Manager_Cycle_004(t *testing.T) {
	assert := assert.New(t)

	// A -> B -> C -> A is a longer cycle
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)
	c, err := mgr.New("node", "c")
	assert.NoError(err)

	// A -> B and B -> C
	a.(*mockInstance).refs = []string{b.Name()}
	b.(*mockInstance).refs = []string{c.Name()}

	// Try to make C -> A — completes the cycle
	_, err = mgr.UpdateResourceInstance(context.Background(), c.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{"dep": a.Name()},
		Apply:      true,
	})
	assert.Error(err)
	assert.Contains(err.Error(), "circular")
}

func Test_Manager_Cycle_005(t *testing.T) {
	assert := assert.New(t)

	// No references at all — no cycle
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)

	_, err = mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{},
		Apply:      true,
	})
	assert.NoError(err)
}

func Test_Manager_Cycle_006(t *testing.T) {
	assert := assert.New(t)

	// Diamond dependency is fine: A -> B, A -> C, B -> D, C -> D
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)
	_, err = mgr.New("node", "c") // c
	assert.NoError(err)
	d, err := mgr.New("node", "d")
	assert.NoError(err)

	// B -> D
	b.(*mockInstance).refs = []string{d.Name()}

	// A -> B — no cycle
	_, err = mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{"dep": b.Name()},
		Apply:      true,
	})
	assert.NoError(err)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - DESTROY ORDERING

func Test_Manager_Destroy_001(t *testing.T) {
	assert := assert.New(t)

	// Cannot destroy B when A depends on B
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	// A depends on B
	a.(*mockInstance).refs = []string{b.Name()}

	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: b.Name()})
	assert.Error(err)
	assert.Contains(err.Error(), "depends on it")
}

func Test_Manager_Destroy_002(t *testing.T) {
	assert := assert.New(t)

	// Destroying a leaf (no dependents) succeeds and returns 1 instance
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	// A depends on B
	a.(*mockInstance).refs = []string{b.Name()}

	// Destroying A (the dependent) succeeds
	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: a.Name()})
	assert.NoError(err)
	assert.Len(resp.Instances, 1)
	assert.Equal(a.Name(), resp.Instances[0].Name)
}

func Test_Manager_Destroy_003(t *testing.T) {
	assert := assert.New(t)

	// Destroy dependent first, then dependency succeeds
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	// A depends on B
	a.(*mockInstance).refs = []string{b.Name()}

	// Destroy A first (the dependent)
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: a.Name()})
	assert.NoError(err)

	// Now destroy B (the dependency) — should succeed
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: b.Name()})
	assert.NoError(err)
}

func Test_Manager_Destroy_004(t *testing.T) {
	assert := assert.New(t)

	// Chain: A -> B -> C — cannot destroy B or C while dependents exist
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)
	c, err := mgr.New("node", "c")
	assert.NoError(err)

	a.(*mockInstance).refs = []string{b.Name()}
	b.(*mockInstance).refs = []string{c.Name()}

	// Cannot destroy C (B depends on it)
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: c.Name()})
	assert.Error(err)
	assert.Contains(err.Error(), "depends on it")

	// Cannot destroy B (A depends on it)
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: b.Name()})
	assert.Error(err)
	assert.Contains(err.Error(), "depends on it")

	// Destroy in correct order: A, B, C
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: a.Name()})
	assert.NoError(err)
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: b.Name()})
	assert.NoError(err)
	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: c.Name()})
	assert.NoError(err)
}

func Test_Manager_Destroy_005(t *testing.T) {
	assert := assert.New(t)

	// Instance with no references — destroy succeeds
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)

	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: a.Name()})
	assert.NoError(err)
	assert.Len(resp.Instances, 1)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - CASCADE DESTROY

func Test_Manager_Cascade_001(t *testing.T) {
	assert := assert.New(t)

	// Cascade: A -> B, destroy B cascades to A first, then B
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	a.(*mockInstance).refs = []string{b.Name()}

	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{
		Name:    b.Name(),
		Cascade: true,
	})
	assert.NoError(err)
	assert.Len(resp.Instances, 2)
	// A must be destroyed before B
	assert.Equal(a.Name(), resp.Instances[0].Name)
	assert.Equal(b.Name(), resp.Instances[1].Name)
}

func Test_Manager_Cascade_002(t *testing.T) {
	assert := assert.New(t)

	// Cascade: A -> B -> C, destroy C cascades through A, B, C
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)
	c, err := mgr.New("node", "c")
	assert.NoError(err)

	a.(*mockInstance).refs = []string{b.Name()}
	b.(*mockInstance).refs = []string{c.Name()}

	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{
		Name:    c.Name(),
		Cascade: true,
	})
	assert.NoError(err)
	assert.Len(resp.Instances, 3)
	// A before B before C
	assert.Equal(a.Name(), resp.Instances[0].Name)
	assert.Equal(b.Name(), resp.Instances[1].Name)
	assert.Equal(c.Name(), resp.Instances[2].Name)
}

func Test_Manager_Cascade_003(t *testing.T) {
	assert := assert.New(t)

	// Cascade on a leaf (no dependents) destroys just that instance
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)

	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{
		Name:    a.Name(),
		Cascade: true,
	})
	assert.NoError(err)
	assert.Len(resp.Instances, 1)
	assert.Equal(a.Name(), resp.Instances[0].Name)
}

func Test_Manager_Cascade_004(t *testing.T) {
	assert := assert.New(t)

	// Cascade does not destroy unrelated instances
	// A -> B, C is independent. Cascade-destroy B removes A and B only.
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)
	c, err := mgr.New("node", "c")
	assert.NoError(err)

	a.(*mockInstance).refs = []string{b.Name()}

	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{
		Name:    b.Name(),
		Cascade: true,
	})
	assert.NoError(err)
	assert.Len(resp.Instances, 2)

	// C should still exist
	_, err = mgr.GetResourceInstance(context.Background(), c.Name())
	assert.NoError(err)
}

func Test_Manager_Cascade_005(t *testing.T) {
	assert := assert.New(t)

	// Diamond: A -> C, B -> C. Cascade-destroy C removes A, B, C.
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)
	c, err := mgr.New("node", "c")
	assert.NoError(err)

	a.(*mockInstance).refs = []string{c.Name()}
	b.(*mockInstance).refs = []string{c.Name()}

	resp, err := mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{
		Name:    c.Name(),
		Cascade: true,
	})
	assert.NoError(err)
	assert.Len(resp.Instances, 3)

	// A and B must come before C (order between A and B is unspecified)
	names := make([]string, len(resp.Instances))
	for i, inst := range resp.Instances {
		names[i] = inst.Name
	}
	assert.Equal(c.Name(), names[2], "C must be last")
	assert.Contains(names[:2], a.Name())
	assert.Contains(names[:2], b.Name())
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - LIFECYCLE & GETTERS

func Test_Manager_Lifecycle_001(t *testing.T) {
	assert := assert.New(t)

	// New returns a manager with expected metadata
	mgr, err := provider.New("myprovider", "my description", "1.2.3")
	assert.NoError(err)
	assert.NotNil(mgr)
	assert.Equal("myprovider", mgr.Name())
	assert.Equal("my description", mgr.Description())
	assert.Equal("1.2.3", mgr.Version())
}

func Test_Manager_Lifecycle_002(t *testing.T) {
	assert := assert.New(t)

	// Resources returns registered resources
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.Len(mgr.Resources(), 0)

	assert.NoError(mgr.RegisterResource(&mockResource{name: "alpha"}))
	assert.NoError(mgr.RegisterResource(&mockResource{name: "beta"}))
	assert.Len(mgr.Resources(), 2)
}

func Test_Manager_Lifecycle_003(t *testing.T) {
	assert := assert.New(t)

	// Close destroys all instances in dependency order
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)
	b, err := mgr.New("node", "b")
	assert.NoError(err)

	// A depends on B
	a.(*mockInstance).refs = []string{b.Name()}

	// Close should destroy everything without error
	err = mgr.Close(context.Background())
	assert.NoError(err)

	// Instances should be gone
	_, err = mgr.GetResourceInstance(context.Background(), a.Name())
	assert.Error(err)
	_, err = mgr.GetResourceInstance(context.Background(), b.Name())
	assert.Error(err)
}

func Test_Manager_Lifecycle_004(t *testing.T) {
	assert := assert.New(t)

	// Close with no instances is fine
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.Close(context.Background()))
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - REGISTER RESOURCE

func Test_Manager_Register_001(t *testing.T) {
	assert := assert.New(t)

	// Registering nil resource returns error
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.Error(mgr.RegisterResource(nil))
}

func Test_Manager_Register_002(t *testing.T) {
	assert := assert.New(t)

	// Registering resource with empty name returns error
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.Error(mgr.RegisterResource(&mockResource{name: ""}))
}

func Test_Manager_Register_003(t *testing.T) {
	assert := assert.New(t)

	// Registering duplicate resource returns error
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))
	assert.Error(mgr.RegisterResource(&mockResource{name: "node"}))
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - REGISTER READONLY INSTANCE

func Test_Manager_RegisterReadonlyInstance_001(t *testing.T) {
	assert := assert.New(t)

	// Happy path: creates, validates, applies, and stores as read-only
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	inst, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "main", nil)
	assert.NoError(err)
	assert.NotNil(inst)
	assert.Equal("node.main", inst.Name())

	// Verify the instance is retrievable and marked read-only
	got, err := mgr.GetResourceInstance(context.Background(), "node.main")
	assert.NoError(err)
	assert.Equal("node.main", got.Instance.Name)
	assert.True(got.Instance.ReadOnly)
}

func Test_Manager_RegisterReadonlyInstance_002(t *testing.T) {
	assert := assert.New(t)

	// Auto-registers the resource type if not already registered
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	inst, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "x", nil)
	assert.NoError(err)
	assert.NotNil(inst)

	// Resource should now be registered and visible
	resp, err := mgr.ListResources(context.Background(), schema.ListResourcesRequest{})
	assert.NoError(err)
	assert.Len(resp.Resources, 1)
	assert.Equal("node", resp.Resources[0].Name)
}

func Test_Manager_RegisterReadonlyInstance_003(t *testing.T) {
	assert := assert.New(t)

	// Instance is visible in ListResources
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "a", nil)
	assert.NoError(err)
	_, err = mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "b", nil)
	assert.NoError(err)

	resp, err := mgr.ListResources(context.Background(), schema.ListResourcesRequest{})
	assert.NoError(err)
	assert.Len(resp.Resources, 1)
	assert.Len(resp.Resources[0].Instances, 2)
}

func Test_Manager_RegisterReadonlyInstance_004(t *testing.T) {
	assert := assert.New(t)

	// Returns the concrete instance so callers can type-assert
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	inst, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "x", nil)
	assert.NoError(err)
	_, ok := inst.(*mockInstance)
	assert.True(ok, "expected *mockInstance, got %T", inst)
}

func Test_Manager_RegisterReadonlyInstance_005(t *testing.T) {
	assert := assert.New(t)

	// Instance is protected: update and destroy are rejected
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	inst, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "x", nil)
	assert.NoError(err)

	_, err = mgr.UpdateResourceInstance(context.Background(), inst.Name(), schema.UpdateResourceInstanceRequest{
		Apply: true,
	})
	assert.Error(err)

	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{
		Name: inst.Name(),
	})
	assert.Error(err)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - LIST RESOURCES

func Test_Manager_List_001(t *testing.T) {
	assert := assert.New(t)

	// List all resources (no filter)
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "alpha"}))
	assert.NoError(mgr.RegisterResource(&mockResource{name: "beta"}))

	resp, err := mgr.ListResources(context.Background(), schema.ListResourcesRequest{})
	assert.NoError(err)
	assert.Equal("test", resp.Provider)
	assert.Equal("test provider", resp.Description)
	assert.Equal("0.0.1", resp.Version)
	assert.Len(resp.Resources, 2)
}

func Test_Manager_List_002(t *testing.T) {
	assert := assert.New(t)

	// List with type filter
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "alpha"}))
	assert.NoError(mgr.RegisterResource(&mockResource{name: "beta"}))

	filter := "alpha"
	resp, err := mgr.ListResources(context.Background(), schema.ListResourcesRequest{Type: &filter})
	assert.NoError(err)
	assert.Len(resp.Resources, 1)
	assert.Equal("alpha", resp.Resources[0].Name)
}

func Test_Manager_List_003(t *testing.T) {
	assert := assert.New(t)

	// List with invalid type filter returns error
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "alpha"}))

	filter := "nosuch"
	_, err = mgr.ListResources(context.Background(), schema.ListResourcesRequest{Type: &filter})
	assert.Error(err)
}

func Test_Manager_List_004(t *testing.T) {
	assert := assert.New(t)

	// List includes instance metadata
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	_, err = mgr.New("node", "a")
	assert.NoError(err)
	_, err = mgr.New("node", "b")
	assert.NoError(err)

	resp, err := mgr.ListResources(context.Background(), schema.ListResourcesRequest{})
	assert.NoError(err)
	assert.Len(resp.Resources, 1)
	assert.Len(resp.Resources[0].Instances, 2)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - CREATE RESOURCE INSTANCE

func Test_Manager_Create_001(t *testing.T) {
	assert := assert.New(t)

	// Create via public API
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	resp, err := mgr.CreateResourceInstance(context.Background(), schema.CreateResourceInstanceRequest{
		Name: "node.x",
	})
	assert.NoError(err)
	assert.Equal("node", resp.Instance.Resource)
	assert.Equal("node.x", resp.Instance.Name)
}

func Test_Manager_Create_002(t *testing.T) {
	assert := assert.New(t)

	// Create with unknown resource type returns error
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.CreateResourceInstance(context.Background(), schema.CreateResourceInstanceRequest{
		Name: "nosuch.x",
	})
	assert.Error(err)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - GET / UPDATE / DESTROY ERROR PATHS

func Test_Manager_Get_001(t *testing.T) {
	assert := assert.New(t)

	// Get non-existent instance returns not found
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.GetResourceInstance(context.Background(), "nosuch")
	assert.Error(err)
	assert.Contains(err.Error(), "not found")
}

func Test_Manager_Update_001(t *testing.T) {
	assert := assert.New(t)

	// Update non-existent instance returns not found
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.UpdateResourceInstance(context.Background(), "nosuch", schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{},
	})
	assert.Error(err)
	assert.Contains(err.Error(), "not found")
}

func Test_Manager_Update_002(t *testing.T) {
	assert := assert.New(t)

	// Plan-only update (Apply=false) returns plan without changing state
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)
	assert.NoError(mgr.RegisterResource(&mockResource{name: "node"}))

	a, err := mgr.New("node", "a")
	assert.NoError(err)

	resp, err := mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{},
		Apply:      false,
	})
	assert.NoError(err)
	assert.NotNil(resp)

	// State should still be nil (nothing applied)
	get, err := mgr.GetResourceInstance(context.Background(), a.Name())
	assert.NoError(err)
	assert.Nil(get.Instance.State)
}

func Test_Manager_Destroy_006(t *testing.T) {
	assert := assert.New(t)

	// Destroy non-existent instance returns not found
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: "nosuch"})
	assert.Error(err)
	assert.Contains(err.Error(), "not found")
}

func Test_Manager_New_001(t *testing.T) {
	assert := assert.New(t)

	// New with unregistered resource type returns error
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.New("nosuch", "a")
	assert.Error(err)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - READ-ONLY (DATA) INSTANCES

func Test_Manager_ReadOnly_001(t *testing.T) {
	assert := assert.New(t)

	// Cannot update a read-only instance
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	a, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "ro1", nil)
	assert.NoError(err)

	_, err = mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{},
		Apply:      true,
	})
	assert.Error(err)
	assert.Contains(err.Error(), "read-only")
}

func Test_Manager_ReadOnly_002(t *testing.T) {
	assert := assert.New(t)

	// Cannot destroy a read-only instance
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	a, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "ro2", nil)
	assert.NoError(err)

	_, err = mgr.DestroyResourceInstance(context.Background(), schema.DestroyResourceInstanceRequest{Name: a.Name()})
	assert.Error(err)
	assert.Contains(err.Error(), "read-only")
}

func Test_Manager_ReadOnly_003(t *testing.T) {
	assert := assert.New(t)

	// Can read a read-only instance
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	a, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "ro3", nil)
	assert.NoError(err)

	resp, err := mgr.GetResourceInstance(context.Background(), a.Name())
	assert.NoError(err)
	assert.Equal(a.Name(), resp.Instance.Name)
	assert.True(resp.Instance.ReadOnly)
}

func Test_Manager_ReadOnly_004(t *testing.T) {
	assert := assert.New(t)

	// ReadOnly appears in list output
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	_, err = mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "ro4", nil)
	assert.NoError(err)

	resp, err := mgr.ListResources(context.Background(), schema.ListResourcesRequest{})
	assert.NoError(err)
	assert.Len(resp.Resources, 1)
	assert.Len(resp.Resources[0].Instances, 1)
	assert.True(resp.Resources[0].Instances[0].ReadOnly)
}

func Test_Manager_ReadOnly_005(t *testing.T) {
	assert := assert.New(t)

	// Plan-only update on read-only instance is also rejected
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	a, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "ro5", nil)
	assert.NoError(err)

	_, err = mgr.UpdateResourceInstance(context.Background(), a.Name(), schema.UpdateResourceInstanceRequest{
		Attributes: schema.State{},
		Apply:      false,
	})
	assert.Error(err)
	assert.Contains(err.Error(), "read-only")
}

func Test_Manager_ReadOnly_006(t *testing.T) {
	assert := assert.New(t)

	// Close still destroys read-only instances (cleanup)
	mgr, err := provider.New("test", "test provider", "0.0.1")
	assert.NoError(err)

	a, err := mgr.RegisterReadonlyInstance(context.Background(), &mockResource{name: "node"}, "ro6", nil)
	assert.NoError(err)

	err = mgr.Close(context.Background())
	assert.NoError(err)

	// Should be gone
	_, err = mgr.GetResourceInstance(context.Background(), a.Name())
	assert.Error(err)
}
