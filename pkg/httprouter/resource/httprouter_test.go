package resource_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	server "github.com/mutablelogic/go-server"
	resource "github.com/mutablelogic/go-server/pkg/httprouter/resource"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	"github.com/mutablelogic/go-server/pkg/provider/schema/schematest"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// MOCK TYPES

// mockHandler implements schema.ResourceInstance and server.HTTPHandler.
type mockHandler struct {
	schematest.ResourceInstance
	path string
}

var _ server.HTTPHandler = (*mockHandler)(nil)

func (m *mockHandler) HandlerPath() string { return m.path }
func (m *mockHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func (m *mockHandler) Spec() *openapi.PathItem { return nil }

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// mockResolver returns a Resolver that returns nil for every name.
func mockResolver() schema.Resolver {
	return func(string) schema.ResourceInstance { return nil }
}

// validState returns a minimal valid State for an httprouter resource.
func validState() schema.State {
	return schema.State{
		"prefix":  "/",
		"title":   "Test API",
		"version": "1.0.0",
		"openapi": true,
	}
}

// newInstance creates a fresh ResourceInstance via New().
func newInstance(t *testing.T) schema.ResourceInstance {
	t.Helper()
	var r resource.Resource
	inst, err := r.New()
	assert.NoError(t, err)
	assert.NotNil(t, inst)
	return inst
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - RESOURCE

func Test_Resource_001(t *testing.T) {
	// Name returns "httprouter"
	assert := assert.New(t)
	var r resource.Resource
	assert.Equal("httprouter", r.Name())
}

func Test_Resource_002(t *testing.T) {
	// Schema returns 9 attributes
	assert := assert.New(t)
	var r resource.Resource
	attrs := r.Schema()
	assert.Len(attrs, 8)

	names := make(map[string]bool, len(attrs))
	for _, a := range attrs {
		names[a.Name] = true
	}
	for _, want := range []string{"prefix", "origin", "title", "version", "endpoints", "openapi", "middleware", "handlers"} {
		assert.True(names[want], "missing attribute %q", want)
	}
}

func Test_Resource_003(t *testing.T) {
	// New creates instances with empty names (manager assigns type.label)
	assert := assert.New(t)
	var r resource.Resource
	inst1, err := r.New()
	assert.NoError(err)
	assert.Empty(inst1.Name())

	inst2, err := r.New()
	assert.NoError(err)
	assert.Empty(inst2.Name())
}

func Test_Resource_004(t *testing.T) {
	// Resource() on an instance returns the httprouter resource type
	assert := assert.New(t)
	inst := newInstance(t)
	assert.Equal("httprouter", inst.Resource().Name())
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - VALIDATE

func Test_Validate_001(t *testing.T) {
	// Valid config - no error
	assert := assert.New(t)
	inst := newInstance(t)
	config, err := inst.Validate(context.Background(), validState(), mockResolver())
	assert.NoError(err)
	assert.NotNil(config)
}

func Test_Validate_002(t *testing.T) {
	// Missing required title - error
	assert := assert.New(t)
	inst := newInstance(t)
	state := validState()
	delete(state, "title")
	_, err := inst.Validate(context.Background(), state, mockResolver())
	assert.Error(err)
	assert.Contains(err.Error(), "title")
}

func Test_Validate_003(t *testing.T) {
	// Missing required version - error
	assert := assert.New(t)
	inst := newInstance(t)
	state := validState()
	delete(state, "version")
	_, err := inst.Validate(context.Background(), state, mockResolver())
	assert.Error(err)
	assert.Contains(err.Error(), "version")
}

func Test_Validate_004(t *testing.T) {
	// Empty title string - error
	assert := assert.New(t)
	inst := newInstance(t)
	state := validState()
	state["title"] = ""
	_, err := inst.Validate(context.Background(), state, mockResolver())
	assert.Error(err)
	assert.Contains(err.Error(), "title")
}

func Test_Validate_005(t *testing.T) {
	// Empty version string - error
	assert := assert.New(t)
	inst := newInstance(t)
	state := validState()
	state["version"] = ""
	_, err := inst.Validate(context.Background(), state, mockResolver())
	assert.Error(err)
	assert.Contains(err.Error(), "version")
}

func Test_Validate_006(t *testing.T) {
	// Prefix is normalised (no trailing slash error)
	assert := assert.New(t)
	inst := newInstance(t)
	state := validState()
	state["prefix"] = "/api/v1/"
	config, err := inst.Validate(context.Background(), state, mockResolver())
	assert.NoError(err)
	assert.NotNil(config)
}

func Test_Validate_007(t *testing.T) {
	// Minimal valid state — only required fields
	assert := assert.New(t)
	inst := newInstance(t)
	state := schema.State{
		"title":   "My API",
		"version": "2.0.0",
	}
	config, err := inst.Validate(context.Background(), state, mockResolver())
	assert.NoError(err)
	assert.NotNil(config)
	c := config.(*resource.Resource)
	assert.Equal("My API", c.Title)
	assert.Equal("2.0.0", c.Version)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - PLAN

func Test_Plan_001(t *testing.T) {
	// Plan on a fresh instance produces ActionCreate
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
	}
	plan, err := inst.Plan(context.Background(), config)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
	assert.NotEmpty(plan.Changes)
}

func Test_Plan_002(t *testing.T) {
	// Plan with matching config after Apply produces ActionNoop
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))

	plan, err := inst.Plan(context.Background(), config)
	assert.NoError(err)
	assert.Equal(schema.ActionNoop, plan.Action)
	assert.Empty(plan.Changes)
}

func Test_Plan_003(t *testing.T) {
	// Plan with different title after Apply produces ActionUpdate
	assert := assert.New(t)
	inst := newInstance(t)
	old := &resource.Resource{
		Prefix:  "/",
		Title:   "Old Title",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), old))

	updated := &resource.Resource{
		Prefix:  "/",
		Title:   "New Title",
		Version: "1.0.0",
		OpenAPI: true,
	}
	plan, err := inst.Plan(context.Background(), updated)
	assert.NoError(err)
	assert.Equal(schema.ActionUpdate, plan.Action)

	found := false
	for _, ch := range plan.Changes {
		if ch.Field == "title" {
			found = true
			assert.Equal("Old Title", ch.Old)
			assert.Equal("New Title", ch.New)
		}
	}
	assert.True(found, "expected a change for 'title'")
}

func Test_Plan_004(t *testing.T) {
	// Plan on a fresh instance is always ActionCreate
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/api",
		Title:   "API",
		Version: "0.1.0",
	}
	plan, err := inst.Plan(context.Background(), config)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - APPLY

func Test_Apply_001(t *testing.T) {
	// Apply succeeds with valid config
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))
}

func Test_Apply_002(t *testing.T) {
	// Apply twice replaces the router (no error)
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))
	assert.NoError(inst.Apply(context.Background(), config))
}

func Test_Apply_003(t *testing.T) {
	// Apply with wrong type returns error (not panic)
	assert := assert.New(t)
	inst := newInstance(t)
	err := inst.Apply(context.Background(), "not a *Resource")
	assert.Error(err)
	assert.Contains(err.Error(), "unexpected config type")
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - SERVE HTTP

func Test_ServeHTTP_001(t *testing.T) {
	// Before Apply, ServeHTTP returns 503 Service Unavailable
	assert := assert.New(t)
	inst := newInstance(t)
	handler := inst.(http.Handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)
}

func Test_ServeHTTP_002(t *testing.T) {
	// After Apply, unknown path returns JSON 404
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))

	handler := inst.(http.Handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(http.StatusNotFound, rec.Code)
	assert.Contains(rec.Header().Get("Content-Type"), "application/json")
}

func Test_ServeHTTP_003(t *testing.T) {
	// After Apply with OpenAPI enabled, /openapi.json returns the spec
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "My API",
		Version: "2.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))

	handler := inst.(http.Handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(http.StatusOK, rec.Code)

	var spec map[string]any
	assert.NoError(json.Unmarshal(rec.Body.Bytes(), &spec))
	assert.Equal("3.1.1", spec["openapi"])
	info := spec["info"].(map[string]any)
	assert.Equal("My API", info["title"])
	assert.Equal("2.0.0", info["version"])
}

func Test_ServeHTTP_004(t *testing.T) {
	// With OpenAPI disabled, /openapi.json returns 404
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: false,
	}
	assert.NoError(inst.Apply(context.Background(), config))

	handler := inst.(http.Handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(http.StatusNotFound, rec.Code)
}

func Test_ServeHTTP_005(t *testing.T) {
	// Unknown paths return JSON 404 from the default handler
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: false,
	}
	assert.NoError(inst.Apply(context.Background(), config))

	handler := inst.(http.Handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(http.StatusNotFound, rec.Code)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - DESTROY

func Test_Destroy_001(t *testing.T) {
	// Destroy on a never-applied instance is a no-op
	assert := assert.New(t)
	inst := newInstance(t)
	assert.NoError(inst.Destroy(context.Background()))
}

func Test_Destroy_002(t *testing.T) {
	// After Destroy, ServeHTTP returns 503 again
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))
	assert.NoError(inst.Destroy(context.Background()))

	handler := inst.(http.Handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - REFERENCES

func Test_References_001(t *testing.T) {
	// References before Apply returns nil (no references for httprouter)
	assert := assert.New(t)
	inst := newInstance(t)
	assert.Nil(inst.References())
}

func Test_References_002(t *testing.T) {
	// References after Apply still returns nil (httprouter has no reference fields)
	assert := assert.New(t)
	inst := newInstance(t)
	config := &resource.Resource{
		Prefix:  "/",
		Title:   "Test",
		Version: "1.0.0",
		OpenAPI: true,
	}
	assert.NoError(inst.Apply(context.Background(), config))
	assert.Nil(inst.References())
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - DUPLICATE HANDLER PATHS

func Test_Apply_004(t *testing.T) {
	// Apply with duplicate handler paths returns error
	assert := assert.New(t)
	inst := newInstance(t)

	h1 := &mockHandler{
		ResourceInstance: schematest.ResourceInstance{N: "handler-1", RN: "httphandler"},
		path:             "resource/{id}",
	}
	h2 := &mockHandler{
		ResourceInstance: schematest.ResourceInstance{N: "handler-2", RN: "httphandler"},
		path:             "resource/{id}",
	}

	config := &resource.Resource{
		Prefix:   "/",
		Title:    "Test",
		Version:  "1.0.0",
		Handlers: []schema.ResourceInstance{h1, h2},
	}
	err := inst.Apply(context.Background(), config)
	assert.Error(err)
	assert.Contains(err.Error(), "duplicate handler path")
	assert.Contains(err.Error(), "resource/{id}")
}

func Test_Apply_005(t *testing.T) {
	// Apply with distinct handler paths succeeds
	assert := assert.New(t)
	inst := newInstance(t)

	h1 := &mockHandler{
		ResourceInstance: schematest.ResourceInstance{N: "handler-1", RN: "httphandler"},
		path:             "resource",
	}
	h2 := &mockHandler{
		ResourceInstance: schematest.ResourceInstance{N: "handler-2", RN: "httphandler"},
		path:             "resource/{id}",
	}

	config := &resource.Resource{
		Prefix:   "/",
		Title:    "Test",
		Version:  "1.0.0",
		Handlers: []schema.ResourceInstance{h1, h2},
	}
	assert.NoError(inst.Apply(context.Background(), config))
}
