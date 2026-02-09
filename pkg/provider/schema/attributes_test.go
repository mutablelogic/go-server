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
	attrs := schema.Attributes(struct{}{})
	assert.Empty(attrs)
}

func Test_Attributes_002(t *testing.T) {
	assert := assert.New(t)
	type simple struct {
		Skipped string `name:"-"`
		Name    string `name:"name" help:"The name"`
	}
	attrs := schema.Attributes(simple{})
	assert.Len(attrs, 1)
	assert.Equal("name", attrs[0].Name)
	assert.Equal("string", attrs[0].Type)
	assert.Equal("The name", attrs[0].Description)
}

func Test_Attributes_003(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.Attributes(httpserverResource{})
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
	attrs := schema.Attributes(httpserverResource{})
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
	attrs := schema.Attributes(httpserverResource{})
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
	attrs := schema.Attributes(httpserverResource{})
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
	attrs := schema.Attributes(&httpserverResource{})
	assert.Len(attrs, 8)
}

func Test_Attributes_008(t *testing.T) {
	assert := assert.New(t)
	attrs := schema.Attributes("not a struct")
	assert.Nil(attrs)
}

func Test_Attributes_009(t *testing.T) {
	assert := assert.New(t)

	// Sensitive and required tags
	type secrets struct {
		Password string `name:"password" help:"The password" sensitive:"" required:""`
		APIKey   string `name:"api-key" sensitive:""`
	}
	attrs := schema.Attributes(secrets{})
	assert.Len(attrs, 2)

	assert.Equal("password", attrs[0].Name)
	assert.True(attrs[0].Sensitive)
	assert.True(attrs[0].Required)

	assert.Equal("api-key", attrs[1].Name)
	assert.True(attrs[1].Sensitive)
	assert.False(attrs[1].Required)
}

///////////////////////////////////////////////////////////////////////////////
// StateOf tests

func Test_StateOf_001(t *testing.T) {
	assert := assert.New(t)
	s := schema.StateOf(struct{}{})
	assert.Empty(s)
}

func Test_StateOf_002(t *testing.T) {
	assert := assert.New(t)
	s := schema.StateOf("not a struct")
	assert.Nil(s)
}

func Test_StateOf_003(t *testing.T) {
	assert := assert.New(t)
	r := httpserverResource{
		Listen:       "localhost:8080",
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Second,
	}
	r.TLS.Name = "myserver"
	r.TLS.Verify = true
	r.TLS.Cert = "/tmp/cert.pem"
	r.TLS.Key = "/tmp/key.pem"

	s := schema.StateOf(r)

	// Router (interface) should be skipped
	_, hasRouter := s["router"]
	assert.False(hasRouter)

	// Scalar fields
	assert.Equal("localhost:8080", s["listen"])
	assert.Equal("5m0s", s["read-timeout"])
	assert.Equal("10s", s["write-timeout"])

	// Embedded TLS fields with prefix
	assert.Equal("myserver", s["tls.name"])
	assert.Equal(true, s["tls.verify"])
	assert.Equal("/tmp/cert.pem", s["tls.cert"])
	assert.Equal("/tmp/key.pem", s["tls.key"])
}

func Test_StateOf_004(t *testing.T) {
	assert := assert.New(t)

	// Pointer input should work the same
	r := &httpserverResource{
		Listen: ":9090",
	}
	s := schema.StateOf(r)
	assert.Equal(":9090", s["listen"])
}

func Test_StateOf_005(t *testing.T) {
	assert := assert.New(t)

	// Fields with name:"-" should be skipped
	type withSkip struct {
		Hidden  string `name:"-"`
		Visible string `name:"visible"`
	}
	s := schema.StateOf(withSkip{Visible: "hello"})
	_, hasHidden := s["hidden"]
	assert.False(hasHidden)
	assert.Equal("hello", s["visible"])
}

///////////////////////////////////////////////////////////////////////////////
// ValidateRefs tests

// mockResourceInstance is a minimal ResourceInstance for testing.
type mockResourceInstance struct {
	name         string
	resourceName string
}

func (m *mockResourceInstance) Name() string                     { return m.name }
func (m *mockResourceInstance) Resource() schema.Resource        { return &mockResource{name: m.resourceName} }
func (m *mockResourceInstance) Validate(_ context.Context) error { return nil }
func (m *mockResourceInstance) Plan(_ context.Context, _ schema.State) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockResourceInstance) Apply(_ context.Context, _ schema.State) (schema.State, error) {
	return nil, nil
}
func (m *mockResourceInstance) Destroy(_ context.Context, _ schema.State) error { return nil }
func (m *mockResourceInstance) References() []string                            { return nil }

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
