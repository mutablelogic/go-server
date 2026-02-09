package resource_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	// Packages
	resource "github.com/mutablelogic/go-server/pkg/httpserver/resource"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// MOCK TYPES

// mockRouter implements both schema.ResourceInstance and http.Handler.
type mockRouter struct {
	name         string
	resourceName string
}

func (m *mockRouter) Name() string             { return m.name }
func (m *mockRouter) Resource() schema.Resource { return &mockResource{name: m.resourceName} }
func (m *mockRouter) Validate(_ context.Context) error { return nil }
func (m *mockRouter) Plan(_ context.Context, _ schema.State) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockRouter) Apply(_ context.Context, _ schema.State) (schema.State, error) {
	return nil, nil
}
func (m *mockRouter) Destroy(_ context.Context, _ schema.State) error { return nil }
func (m *mockRouter) References() []string                            { return nil }
func (m *mockRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// mockResource is a minimal schema.Resource for the mock router.
type mockResource struct {
	name string
}

func (m *mockResource) Name() string                          { return m.name }
func (m *mockResource) Schema() []schema.Attribute            { return nil }
func (m *mockResource) New() (schema.ResourceInstance, error) { return nil, nil }

// mockNonHandler implements schema.ResourceInstance but NOT http.Handler.
type mockNonHandler struct {
	name         string
	resourceName string
}

func (m *mockNonHandler) Name() string             { return m.name }
func (m *mockNonHandler) Resource() schema.Resource { return &mockResource{name: m.resourceName} }
func (m *mockNonHandler) Validate(_ context.Context) error { return nil }
func (m *mockNonHandler) Plan(_ context.Context, _ schema.State) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockNonHandler) Apply(_ context.Context, _ schema.State) (schema.State, error) {
	return nil, nil
}
func (m *mockNonHandler) Destroy(_ context.Context, _ schema.State) error { return nil }
func (m *mockNonHandler) References() []string                            { return nil }

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// freePort returns a TCP port that is available for listening.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	_, port, _ := net.SplitHostPort(l.Addr().String())
	return fmt.Sprintf("localhost:%s", port)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - RESOURCE

func Test_Resource_001(t *testing.T) {
	// Name returns "httpserver"
	assert := assert.New(t)
	var r resource.Resource
	assert.Equal("httpserver", r.Name())
}

func Test_Resource_002(t *testing.T) {
	// Schema returns 8 attributes
	assert := assert.New(t)
	var r resource.Resource
	attrs := r.Schema()
	assert.Len(attrs, 8)

	names := make(map[string]bool, len(attrs))
	for _, a := range attrs {
		names[a.Name] = true
	}
	for _, want := range []string{"listen", "router", "read-timeout", "write-timeout", "tls.name", "tls.verify", "tls.cert", "tls.key"} {
		assert.True(names[want], "missing attribute %q", want)
	}
}

func Test_Resource_003(t *testing.T) {
	// New creates a ResourceInstance with unique counter-based names
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	inst1, err := r.New()
	assert.NoError(err)
	assert.NotNil(inst1)
	assert.Contains(inst1.Name(), "httpserver-")

	inst2, err := r.New()
	assert.NoError(err)
	assert.NotEqual(inst1.Name(), inst2.Name(), "names must be unique")
}

func Test_Resource_004(t *testing.T) {
	// Resource() on an instance returns the config
	assert := assert.New(t)
	r := resource.Resource{Listen: "localhost:8080"}
	inst, err := r.New()
	assert.NoError(err)
	assert.Equal("httpserver", inst.Resource().Name())
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - VALIDATE

func Test_Validate_001(t *testing.T) {
	// Valid config - no error
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)
	assert.NoError(inst.Validate(context.Background()))
}

func Test_Validate_002(t *testing.T) {
	// Missing required router - error
	assert := assert.New(t)
	r := resource.Resource{Listen: "localhost:8080"}
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "router")
}

func Test_Validate_003(t *testing.T) {
	// Wrong router type - error
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "wrongtype"},
	}
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "httprouter")
}

func Test_Validate_004(t *testing.T) {
	// Negative read timeout - error
	assert := assert.New(t)
	r := resource.Resource{
		Listen:      "localhost:8080",
		Router:      &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout: -1 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "read timeout")
}

func Test_Validate_005(t *testing.T) {
	// Negative write timeout - error
	assert := assert.New(t)
	r := resource.Resource{
		Listen:       "localhost:8080",
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		WriteTimeout: -1 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "write timeout")
}

func Test_Validate_006(t *testing.T) {
	// TLS cert without key - error
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = "/some/cert.pem"
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "tls.cert and tls.key must both be set")
}

func Test_Validate_007(t *testing.T) {
	// TLS key without cert - error
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Key = "/some/key.pem"
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "tls.cert and tls.key must both be set")
}

func Test_Validate_008(t *testing.T) {
	// TLS cert file does not exist - error
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = "/nonexistent/cert.pem"
	r.TLS.Key = "/nonexistent/key.pem"
	inst, err := r.New()
	assert.NoError(err)
	err = inst.Validate(context.Background())
	assert.Error(err)
	assert.Contains(err.Error(), "tls.cert")
}

func Test_Validate_009(t *testing.T) {
	// Zero timeouts are fine (no error)
	assert := assert.New(t)
	r := resource.Resource{
		Listen:       "localhost:8080",
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  0,
		WriteTimeout: 0,
	}
	inst, err := r.New()
	assert.NoError(err)
	assert.NoError(inst.Validate(context.Background()))
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - PLAN

func Test_Plan_001(t *testing.T) {
	// Plan with nil current state produces ActionCreate
	assert := assert.New(t)
	r := resource.Resource{
		Listen:      "localhost:8080",
		Router:      &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout: 5 * time.Minute,
	}
	inst, err := r.New()
	assert.NoError(err)
	plan, err := inst.Plan(context.Background(), nil)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
	assert.NotEmpty(plan.Changes)
}

func Test_Plan_002(t *testing.T) {
	// Plan with matching current state produces ActionNoop
	assert := assert.New(t)
	r := resource.Resource{
		Listen:       "localhost:8080",
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	inst, err := r.New()
	assert.NoError(err)

	current := schema.StateOf(r)
	plan, err := inst.Plan(context.Background(), current)
	assert.NoError(err)
	assert.Equal(schema.ActionNoop, plan.Action)
	assert.Empty(plan.Changes)
}

func Test_Plan_003(t *testing.T) {
	// Plan with different state produces ActionUpdate with changes
	assert := assert.New(t)
	r := resource.Resource{
		Listen:       "localhost:8080",
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	inst, err := r.New()
	assert.NoError(err)

	// Build a "current" state that differs in listen
	current := schema.StateOf(r)
	current["listen"] = "localhost:9090"

	plan, err := inst.Plan(context.Background(), current)
	assert.NoError(err)
	assert.Equal(schema.ActionUpdate, plan.Action)
	assert.NotEmpty(plan.Changes)

	found := false
	for _, ch := range plan.Changes {
		if ch.Field == "listen" {
			found = true
			assert.Equal("localhost:9090", ch.Old)
			assert.Equal("localhost:8080", ch.New)
		}
	}
	assert.True(found, "expected a change for 'listen'")
}

func Test_Plan_004(t *testing.T) {
	// Plan with empty current state produces ActionCreate (same as nil)
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)
	plan, err := inst.Plan(context.Background(), schema.State{})
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - APPLY / DESTROY

func Test_Apply_001(t *testing.T) {
	// Apply starts a server that can be destroyed
	assert := assert.New(t)
	addr := freePort(t)
	r := resource.Resource{
		Listen:       addr,
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)

	ctx := context.Background()
	state, err := inst.Apply(ctx, nil)
	assert.NoError(err)
	assert.NotNil(state)
	assert.Equal(addr, state["listen"])

	// Give the server a moment to start listening
	time.Sleep(50 * time.Millisecond)

	// Verify the server is accepting connections
	resp, err := http.Get(fmt.Sprintf("http://%s/", addr))
	assert.NoError(err)
	if resp != nil {
		resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)
	}

	// Destroy should shut it down cleanly
	assert.NoError(inst.Destroy(ctx, state))

	// After destroy the port should stop responding
	time.Sleep(50 * time.Millisecond)
	_, err = http.Get(fmt.Sprintf("http://%s/", addr))
	assert.Error(err)
}

func Test_Apply_002(t *testing.T) {
	// Apply with a non-http.Handler router produces error
	assert := assert.New(t)
	r := resource.Resource{
		Listen: freePort(t),
		Router: &mockNonHandler{name: "r1", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)

	_, err = inst.Apply(context.Background(), nil)
	assert.Error(err)
	assert.Contains(err.Error(), "http.Handler")
}

func Test_Apply_003(t *testing.T) {
	// Apply twice replaces the running server (no leak)
	assert := assert.New(t)
	addr := freePort(t)
	r := resource.Resource{
		Listen:       addr,
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)

	ctx := context.Background()

	// First apply
	state1, err := inst.Apply(ctx, nil)
	assert.NoError(err)
	assert.NotNil(state1)
	time.Sleep(50 * time.Millisecond)

	// Second apply on same port should stop-then-start
	state2, err := inst.Apply(ctx, state1)
	assert.NoError(err)
	assert.NotNil(state2)
	time.Sleep(50 * time.Millisecond)

	// Server should still be reachable
	resp, err := http.Get(fmt.Sprintf("http://%s/", addr))
	assert.NoError(err)
	if resp != nil {
		resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)
	}

	// Clean up
	assert.NoError(inst.Destroy(ctx, state2))
}

func Test_Destroy_001(t *testing.T) {
	// Destroy on a never-applied instance is a no-op
	assert := assert.New(t)
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)
	assert.NoError(inst.Destroy(context.Background(), nil))
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - REFERENCES

func Test_References_001(t *testing.T) {
	// References with a router set
	assert := assert.New(t)
	r := resource.Resource{
		Router: &mockRouter{name: "my-router", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)
	refs := inst.References()
	assert.Equal([]string{"my-router"}, refs)
}

func Test_References_002(t *testing.T) {
	// References with nil router
	assert := assert.New(t)
	r := resource.Resource{}
	inst, err := r.New()
	assert.NoError(err)
	assert.Nil(inst.References())
}
