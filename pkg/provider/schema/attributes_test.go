package schema_test

import (
	"context"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider/schema"
	"github.com/stretchr/testify/assert"
)

type httpserverResource struct {
	Listen       string                  `name:"listen" help:"Listen address (e.g. localhost:8080)"`
	Router       schema.ResourceInstance `name:"router" type:"httprouter" required:"" help:"HTTP router"`
	ReadTimeout  time.Duration           `name:"read-timeout" default:"5m" help:"Read timeout"`
	WriteTimeout time.Duration           `name:"write-timeout" default:"5m" help:"Write timeout"`
	TLS          struct {
		Name   string `name:"name" help:"TLS server name"`
		Verify bool   `name:"verify" default:"true" help:"Verify client certificates"`
		Cert   string `name:"cert" type:"file" default:"" help:"TLS certificate PEM file"`
		Key    string `name:"key" type:"file" default:"" help:"TLS key PEM file"`
	} `embed:"" prefix:"tls."`
}

func Test_Attributes_001(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf(struct{}{})
	assert.Empty(attrs)
}

func Test_Attributes_002(t *testing.T) {
	assert := assert.New(t)
	type simple struct {
		Skipped string `name:"-"`
		Name    string `name:"name" help:"The name"`
	}
	attrs := schema.AttributesOf(simple{})
	assert.Len(attrs, 1)
	assert.Equal("name", attrs[0].Name)
	assert.Equal("string", attrs[0].Type)
	assert.Equal("The name", attrs[0].Description)
}

func Test_Attributes_003(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf(httpserverResource{})
	assert.Len(attrs, 8)
	names := make([]string, 0, len(attrs))
	for _, a := range attrs {
		names = append(names, a.Name)
	}
	assert.Contains(names, "listen")
	assert.Contains(names, "router")
	assert.Contains(names, "read-timeout")
	assert.Contains(names, "write-timeout")
	assert.Contains(names, "tls.name")
	assert.Contains(names, "tls.verify")
	assert.Contains(names, "tls.cert")
	assert.Contains(names, "tls.key")
}

func Test_Attributes_004(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf(httpserverResource{})
	var readTimeout *schema.Attribute
	for _, a := range attrs {
		if a.Name == "read-timeout" {
			readTimeout = &a
			break
		}
	}
	if assert.NotNil(readTimeout) {
		assert.Equal("duration", readTimeout.Type)
		assert.Equal("Read timeout", readTimeout.Description)
		assert.Equal("5m", readTimeout.Default)
	}
}

func Test_Attributes_005(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf(httpserverResource{})
	var cert *schema.Attribute
	for _, a := range attrs {
		if a.Name == "tls.cert" {
			cert = &a
			break
		}
	}
	if assert.NotNil(cert) {
		assert.Equal("file", cert.Type)
		assert.Equal("TLS certificate PEM file", cert.Description)
	}
}

func Test_Attributes_006(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf(httpserverResource{})
	var verify *schema.Attribute
	for _, a := range attrs {
		if a.Name == "tls.verify" {
			verify = &a
			break
		}
	}
	if assert.NotNil(verify) {
		assert.Equal("bool", verify.Type)
		assert.Equal("true", verify.Default)
	}
}

func Test_Attributes_007(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf(&httpserverResource{})
	assert.Len(attrs, 8)
}

func Test_Attributes_008(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.AttributesOf("not a struct")
	assert.Nil(attrs)
}

func Test_Attributes_009(t *testing.T) {
	assert := assert.New(t)

	// Sensitive and required tags
	type secrets struct {
		Password string `name:"password" help:"The password" sensitive:"" required:""`
		APIKey   string `name:"api-key" sensitive:""`
	}
	attrs := schema.AttributesOf(secrets{})
	assert.Len(attrs, 2)

	assert.Equal("password", attrs[0].Name)
	assert.True(attrs[0].Sensitive)
	assert.True(attrs[0].Required)

	assert.Equal("api-key", attrs[1].Name)
	assert.True(attrs[1].Sensitive)
	assert.False(attrs[1].Required)
}

///////////////////////////////////////////////////////////////////////////////
// ValidateRefs tests

// mockResourceInstance is a minimal ResourceInstance for testing.
type mockResourceInstance struct {
	name         string
	resourceName string
}

func (m *mockResourceInstance) Name() string              { return m.name }
func (m *mockResourceInstance) Resource() schema.Resource { return &mockResource{name: m.resourceName} }
func (m *mockResourceInstance) Validate(_ context.Context, _ schema.State, _ schema.Resolver) (any, error) {
	return nil, nil
}
func (m *mockResourceInstance) Plan(_ context.Context, _ any) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockResourceInstance) Apply(_ context.Context, _ any) error {
	return nil
}
func (m *mockResourceInstance) Destroy(_ context.Context) error { return nil }
func (m *mockResourceInstance) Read(_ context.Context) (schema.State, error) {
	return nil, nil
}
func (m *mockResourceInstance) References() []string { return nil }

type mockResource struct{ name string }

func (m *mockResource) Name() string                          { return m.name }
func (m *mockResource) Schema() []schema.Attribute            { return nil }
func (m *mockResource) New() (schema.ResourceInstance, error) { return nil, nil }

func Test_ValidateRefs_001(t *testing.T) {
	assert := assert.New(t)
	// No interface fields — no errors
	type noRefs struct {
		Name string `name:"name"`
	}
	assert.NoError(schema.ValidateRefs(noRefs{}))
}

func Test_ValidateRefs_002(t *testing.T) {
	assert := assert.New(t)
	// Required reference is nil — error
	type withRequired struct {
		Router schema.ResourceInstance `name:"router" type:"httprouter" required:""`
	}
	err := schema.ValidateRefs(withRequired{})
	assert.Error(err)
	assert.Contains(err.Error(), "router: required")
}

func Test_ValidateRefs_003(t *testing.T) {
	assert := assert.New(t)
	// Required reference set with correct type — no error
	type withRequired struct {
		Router schema.ResourceInstance `name:"router" type:"httprouter" required:""`
	}
	err := schema.ValidateRefs(withRequired{
		Router: &mockResourceInstance{name: "router-01", resourceName: "httprouter"},
	})
	assert.NoError(err)
}

func Test_ValidateRefs_004(t *testing.T) {
	assert := assert.New(t)
	// Required reference set with wrong type — error
	type withRequired struct {
		Router schema.ResourceInstance `name:"router" type:"httprouter" required:""`
	}
	err := schema.ValidateRefs(withRequired{
		Router: &mockResourceInstance{name: "router-01", resourceName: "wrongtype"},
	})
	assert.Error(err)
	assert.Contains(err.Error(), `must be of type "httprouter"`)
	assert.Contains(err.Error(), `got "wrongtype"`)
}

func Test_ValidateRefs_005(t *testing.T) {
	assert := assert.New(t)
	// Optional reference (no required tag) that is nil — no error
	type withOptional struct {
		Router schema.ResourceInstance `name:"router" type:"httprouter"`
	}
	assert.NoError(schema.ValidateRefs(withOptional{}))
}

func Test_ValidateRefs_006(t *testing.T) {
	assert := assert.New(t)
	// Full httpserverResource with router set correctly
	r := httpserverResource{
		Router: &mockResourceInstance{name: "router-01", resourceName: "httprouter"},
	}
	assert.NoError(schema.ValidateRefs(r))
}

func Test_ValidateRefs_007(t *testing.T) {
	assert := assert.New(t)
	// Full httpserverResource with no router — required error
	err := schema.ValidateRefs(httpserverResource{})
	assert.Error(err)
	assert.Contains(err.Error(), "router: required")
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - REFERENCES OF

func Test_ReferencesOf_001(t *testing.T) {
	assert := assert.New(t)
	// No interface fields — nil
	type noRefs struct {
		Name string `name:"name"`
	}
	assert.Nil(schema.ReferencesOf(noRefs{}))
}

func Test_ReferencesOf_002(t *testing.T) {
	assert := assert.New(t)
	// Nil reference — nil result
	type withRef struct {
		Router schema.ResourceInstance `name:"router"`
	}
	assert.Nil(schema.ReferencesOf(withRef{}))
}

func Test_ReferencesOf_003(t *testing.T) {
	assert := assert.New(t)
	// Non-nil reference — returns the instance name
	type withRef struct {
		Router schema.ResourceInstance `name:"router"`
	}
	refs := schema.ReferencesOf(withRef{
		Router: &mockResourceInstance{name: "router-01", resourceName: "httprouter"},
	})
	assert.Equal([]string{"router-01"}, refs)
}

func Test_ReferencesOf_004(t *testing.T) {
	assert := assert.New(t)
	// Full httpserverResource with router
	r := httpserverResource{
		Router: &mockResourceInstance{name: "my-router", resourceName: "httprouter"},
	}
	refs := schema.ReferencesOf(r)
	assert.Equal([]string{"my-router"}, refs)
}

func Test_ReferencesOf_005(t *testing.T) {
	assert := assert.New(t)
	// Full httpserverResource with nil router — nil
	assert.Nil(schema.ReferencesOf(httpserverResource{}))
}

func Test_ReferencesOf_006(t *testing.T) {
	assert := assert.New(t)
	// Pointer to struct works too
	r := &httpserverResource{
		Router: &mockResourceInstance{name: "ptr-router", resourceName: "httprouter"},
	}
	refs := schema.ReferencesOf(r)
	assert.Equal([]string{"ptr-router"}, refs)
}

///////////////////////////////////////////////////////////////////////////////
// ReferencesFromState tests

func Test_ReferencesFromState_001(t *testing.T) {
	assert := assert.New(t)
	// Schema with a reference attribute extracts the name from state
	attrs := schema.AttributesOf(httpserverResource{})
	state := schema.State{"router": "router-01", "listen": ":8080"}
	refs := schema.ReferencesFromState(attrs, state)
	assert.Equal([]string{"router-01"}, refs)
}

func Test_ReferencesFromState_002(t *testing.T) {
	assert := assert.New(t)
	// No reference value in state returns nil
	attrs := schema.AttributesOf(httpserverResource{})
	state := schema.State{"listen": ":8080"}
	assert.Nil(schema.ReferencesFromState(attrs, state))
}

func Test_ReferencesFromState_003(t *testing.T) {
	assert := assert.New(t)
	// Empty state returns nil
	attrs := schema.AttributesOf(httpserverResource{})
	assert.Nil(schema.ReferencesFromState(attrs, schema.State{}))
}

func Test_ReferencesFromState_004(t *testing.T) {
	assert := assert.New(t)
	// Non-reference string values are not returned
	type noRefs struct {
		Listen string `name:"listen"`
		Name   string `name:"name"`
	}
	attrs := schema.AttributesOf(noRefs{})
	state := schema.State{"listen": ":8080", "name": "myserver"}
	assert.Nil(schema.ReferencesFromState(attrs, state))
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - READONLY

func Test_ReadOnly_001(t *testing.T) {
	assert := assert.New(t)
	// readonly tag is reflected in the Attribute
	type res struct {
		Name string `name:"name" help:"The name"`
		ID   string `name:"id" readonly:"" help:"Auto-generated ID"`
	}
	attrs := schema.AttributesOf(res{})
	assert.Len(attrs, 2)

	nameAttr := attrs[0]
	assert.Equal("name", nameAttr.Name)
	assert.False(nameAttr.ReadOnly)

	idAttr := attrs[1]
	assert.Equal("id", idAttr.Name)
	assert.True(idAttr.ReadOnly)
}

func Test_ReadOnly_002(t *testing.T) {
	assert := assert.New(t)
	// Decode skips readonly fields even when state has a value for them
	type res struct {
		Name string `name:"name"`
		ID   string `name:"id" readonly:""`
	}
	state := schema.State{"name": "hello", "id": "should-be-ignored"}
	var r res
	assert.NoError(state.Decode(&r, nil))
	assert.Equal("hello", r.Name)
	assert.Equal("", r.ID, "readonly field should not be set by Decode")
}

func Test_ReadOnly_003(t *testing.T) {
	assert := assert.New(t)
	// StateOf includes readonly fields (they are part of state after Apply)
	type res struct {
		Name string `name:"name"`
		ID   string `name:"id" readonly:""`
	}
	r := res{Name: "hello", ID: "abc-123"}
	state := schema.StateOf(&r)
	assert.Equal("hello", state["name"])
	assert.Equal("abc-123", state["id"])
}

func Test_ReadOnly_004(t *testing.T) {
	assert := assert.New(t)
	// ValidateRequired skips readonly fields
	type res struct {
		Name string `name:"name" required:""`
		ID   string `name:"id" required:"" readonly:""`
	}
	r := res{Name: "hello"}
	err := schema.ValidateRequired(r)
	assert.NoError(err, "readonly+required field should not fail validation when empty")
}

func Test_ReadOnly_005(t *testing.T) {
	assert := assert.New(t)
	// ValidateRequired still catches non-readonly required fields
	type res struct {
		Name string `name:"name" required:""`
		ID   string `name:"id" required:"" readonly:""`
	}
	r := res{}
	err := schema.ValidateRequired(r)
	assert.Error(err)
	assert.Contains(err.Error(), "name")
	assert.NotContains(err.Error(), "id")
}
