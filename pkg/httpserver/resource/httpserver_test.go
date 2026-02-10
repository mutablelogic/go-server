package resource_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
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

func (m *mockRouter) Name() string              { return m.name }
func (m *mockRouter) Resource() schema.Resource { return &mockResource{name: m.resourceName} }
func (m *mockRouter) Validate(_ context.Context, _ schema.State, _ schema.Resolver) (any, error) {
	return nil, nil
}
func (m *mockRouter) Plan(_ context.Context, _ any) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockRouter) Apply(_ context.Context, _ any) error {
	return nil
}
func (m *mockRouter) Destroy(_ context.Context) error { return nil }
func (m *mockRouter) Read(_ context.Context) (schema.State, error) {
	return nil, nil
}
func (m *mockRouter) References() []string { return nil }
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

func (m *mockNonHandler) Name() string              { return m.name }
func (m *mockNonHandler) Resource() schema.Resource { return &mockResource{name: m.resourceName} }
func (m *mockNonHandler) Validate(_ context.Context, _ schema.State, _ schema.Resolver) (any, error) {
	return nil, nil
}
func (m *mockNonHandler) Plan(_ context.Context, _ any) (schema.Plan, error) {
	return schema.Plan{}, nil
}
func (m *mockNonHandler) Apply(_ context.Context, _ any) error {
	return nil
}
func (m *mockNonHandler) Destroy(_ context.Context) error { return nil }
func (m *mockNonHandler) Read(_ context.Context) (schema.State, error) {
	return nil, nil
}
func (m *mockNonHandler) References() []string { return nil }

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

// mockResolver returns a Resolver that looks up instances by name.
func mockResolver(instances ...schema.ResourceInstance) schema.Resolver {
	return func(name string) schema.ResourceInstance {
		for _, inst := range instances {
			if inst.Name() == name {
				return inst
			}
		}
		return nil
	}
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
	// Schema returns 9 attributes
	assert := assert.New(t)
	var r resource.Resource
	attrs := r.Schema()
	assert.Len(attrs, 9)

	names := make(map[string]bool, len(attrs))
	for _, a := range attrs {
		names[a.Name] = true
	}
	for _, want := range []string{"listen", "endpoint", "router", "read-timeout", "write-timeout", "tls.name", "tls.verify", "tls.cert", "tls.key"} {
		assert.True(names[want], "missing attribute %q", want)
	}

	// endpoint must be readonly
	for _, a := range attrs {
		if a.Name == "endpoint" {
			assert.True(a.ReadOnly, "endpoint should be readonly")
		}
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
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: router,
	}
	inst, err := r.New()
	assert.NoError(err)
	config, err := inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.NoError(err)
	assert.NotNil(config)
}

func Test_Validate_002(t *testing.T) {
	// Missing required router - error
	assert := assert.New(t)
	r := resource.Resource{Listen: "localhost:8080"}
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver())
	assert.Error(err)
	assert.Contains(err.Error(), "router")
}

func Test_Validate_003(t *testing.T) {
	// Wrong router type - error
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "wrongtype"}
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: router,
	}
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.Error(err)
	assert.Contains(err.Error(), "httprouter")
}

func Test_Validate_004(t *testing.T) {
	// Negative read timeout - error
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen:      "localhost:8080",
		Router:      router,
		ReadTimeout: -1 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.Error(err)
	assert.Contains(err.Error(), "read timeout")
}

func Test_Validate_005(t *testing.T) {
	// Negative write timeout - error
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen:       "localhost:8080",
		Router:       router,
		WriteTimeout: -1 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.Error(err)
	assert.Contains(err.Error(), "write timeout")
}

func Test_Validate_006(t *testing.T) {
	// TLS cert only passes Validate (checked at Plan time)
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: router,
	}
	r.TLS.Cert = []byte("some-cert-data")
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.NoError(err)
}

func Test_Validate_007(t *testing.T) {
	// TLS key only passes Validate (checked at Plan time)
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: router,
	}
	r.TLS.Key = []byte("some-key-data")
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.NoError(err)
}

func Test_Validate_008(t *testing.T) {
	// Both TLS cert and key provided - no validation error
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen: "localhost:8080",
		Router: router,
	}
	r.TLS.Cert = []byte("some-cert-data")
	r.TLS.Key = []byte("some-key-data")
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.NoError(err)
}

func Test_Validate_009(t *testing.T) {
	// Zero timeouts are fine (no error)
	assert := assert.New(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	r := resource.Resource{
		Listen:       "localhost:8080",
		Router:       router,
		ReadTimeout:  0,
		WriteTimeout: 0,
	}
	inst, err := r.New()
	assert.NoError(err)
	_, err = inst.Validate(context.Background(), schema.StateOf(r), mockResolver(router))
	assert.NoError(err)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - PLAN

func Test_Plan_001(t *testing.T) {
	// Plan on a fresh instance produces ActionCreate
	assert := assert.New(t)
	r := &resource.Resource{
		Listen:      "localhost:8080",
		Router:      &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout: 5 * time.Minute,
	}
	inst, err := r.New()
	assert.NoError(err)
	plan, err := inst.Plan(context.Background(), r)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
	assert.NotEmpty(plan.Changes)
}

func Test_Plan_002(t *testing.T) {
	// Plan with matching config after Apply produces ActionNoop
	assert := assert.New(t)
	addr := freePort(t)
	r := &resource.Resource{
		Listen:       addr,
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	inst, err := r.New()
	assert.NoError(err)

	// Apply to establish current state
	assert.NoError(inst.Apply(context.Background(), r))
	defer inst.Destroy(context.Background())

	// Plan with the same config
	plan, err := inst.Plan(context.Background(), r)
	assert.NoError(err)
	assert.Equal(schema.ActionNoop, plan.Action)
	assert.Empty(plan.Changes)
}

func Test_Plan_003(t *testing.T) {
	// Plan with different config after Apply produces ActionUpdate
	assert := assert.New(t)
	addr := freePort(t)
	router := &mockRouter{name: "r1", resourceName: "httprouter"}
	old := &resource.Resource{
		Listen:       addr,
		Router:       router,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	inst, err := old.New()
	assert.NoError(err)

	// Apply the original config
	assert.NoError(inst.Apply(context.Background(), old))
	defer inst.Destroy(context.Background())

	// Plan with a different listen address
	updated := &resource.Resource{
		Listen:       "localhost:9090",
		Router:       router,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	plan, err := inst.Plan(context.Background(), updated)
	assert.NoError(err)
	assert.Equal(schema.ActionUpdate, plan.Action)
	assert.NotEmpty(plan.Changes)

	found := false
	for _, ch := range plan.Changes {
		if ch.Field == "listen" {
			found = true
			assert.Equal(addr, ch.Old)
			assert.Equal("localhost:9090", ch.New)
		}
	}
	assert.True(found, "expected a change for 'listen'")
}

func Test_Plan_004(t *testing.T) {
	// Plan on a fresh instance always produces ActionCreate
	assert := assert.New(t)
	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)
	plan, err := inst.Plan(context.Background(), r)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - APPLY / DESTROY

func Test_Apply_001(t *testing.T) {
	// Apply starts a server that can be destroyed
	assert := assert.New(t)
	addr := freePort(t)
	r := &resource.Resource{
		Listen:       addr,
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)

	ctx := context.Background()
	assert.NoError(inst.Apply(ctx, r))

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
	assert.NoError(inst.Destroy(ctx))

	// After destroy the port should stop responding
	time.Sleep(50 * time.Millisecond)
	_, err = http.Get(fmt.Sprintf("http://%s/", addr))
	assert.Error(err)
}

func Test_Apply_002(t *testing.T) {
	// Apply with a non-http.Handler router produces error
	assert := assert.New(t)
	r := &resource.Resource{
		Listen: freePort(t),
		Router: &mockNonHandler{name: "r1", resourceName: "httprouter"},
	}
	inst, err := r.New()
	assert.NoError(err)

	err = inst.Apply(context.Background(), r)
	assert.Error(err)
	assert.Contains(err.Error(), "http.Handler")
}

func Test_Apply_003(t *testing.T) {
	// Apply twice replaces the running server (no leak)
	assert := assert.New(t)
	addr := freePort(t)
	r := &resource.Resource{
		Listen:       addr,
		Router:       &mockRouter{name: "r1", resourceName: "httprouter"},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)

	ctx := context.Background()

	// First apply
	assert.NoError(inst.Apply(ctx, r))
	time.Sleep(50 * time.Millisecond)

	// Second apply on same port should stop-then-start
	assert.NoError(inst.Apply(ctx, r))
	time.Sleep(50 * time.Millisecond)

	// Server should still be reachable
	resp, err := http.Get(fmt.Sprintf("http://%s/", addr))
	assert.NoError(err)
	if resp != nil {
		resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)
	}

	// Clean up
	assert.NoError(inst.Destroy(ctx))
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
	assert.NoError(inst.Destroy(context.Background()))
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - REFERENCES

func Test_References_001(t *testing.T) {
	// References returns router name after Apply
	assert := assert.New(t)
	addr := freePort(t)
	r := &resource.Resource{
		Listen:       addr,
		Router:       &mockRouter{name: "my-router", resourceName: "httprouter"},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	inst, err := r.New()
	assert.NoError(err)
	assert.NoError(inst.Apply(context.Background(), r))
	defer inst.Destroy(context.Background())
	refs := inst.References()
	assert.Equal([]string{"my-router"}, refs)
}

func Test_References_002(t *testing.T) {
	// References before Apply returns nil
	assert := assert.New(t)
	r := resource.Resource{}
	inst, err := r.New()
	assert.NoError(err)
	assert.Nil(inst.References())
}

///////////////////////////////////////////////////////////////////////////////
// HELPERS - TLS CERTIFICATE GENERATION

// generateTestCert creates a self-signed ECDSA certificate and key pair
// as PEM-encoded byte slices. The certificate is valid from notBefore
// to notAfter.
func generateTestCert(t *testing.T, notBefore, notAfter time.Time) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	return
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - PLAN TLS VALIDATION

func Test_Plan_TLS_001(t *testing.T) {
	// Plan with invalid PEM data - error
	assert := assert.New(t)
	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = []byte("not valid pem")
	r.TLS.Key = []byte("not valid pem")
	inst, err := r.New()
	assert.NoError(err)

	_, err = inst.Plan(context.Background(), r)
	assert.Error(err)
	assert.Contains(err.Error(), "tls")
}

func Test_Plan_TLS_002(t *testing.T) {
	// Plan with expired certificate - error
	assert := assert.New(t)
	expired := time.Now().Add(-48 * time.Hour)
	certPEM, keyPEM := generateTestCert(t, expired.Add(-24*time.Hour), expired)

	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = certPEM
	r.TLS.Key = keyPEM
	inst, err := r.New()
	assert.NoError(err)

	_, err = inst.Plan(context.Background(), r)
	assert.Error(err)
	assert.Contains(err.Error(), "expired")
}

func Test_Plan_TLS_003(t *testing.T) {
	// Plan with valid certificate - no error
	assert := assert.New(t)
	now := time.Now()
	certPEM, keyPEM := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))

	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = certPEM
	r.TLS.Key = keyPEM
	inst, err := r.New()
	assert.NoError(err)

	plan, err := inst.Plan(context.Background(), r)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
}

func Test_Plan_TLS_004(t *testing.T) {
	// Plan with mismatched cert and key - error
	assert := assert.New(t)
	now := time.Now()
	certPEM, _ := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))
	_, keyPEM := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))

	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = certPEM
	r.TLS.Key = keyPEM
	inst, err := r.New()
	assert.NoError(err)

	_, err = inst.Plan(context.Background(), r)
	assert.Error(err)
	assert.Contains(err.Error(), "tls")
}

func Test_Plan_TLS_005(t *testing.T) {
	// Plan with cert+key concatenated in Cert field, Key nil - no error
	assert := assert.New(t)
	now := time.Now()
	certPEM, keyPEM := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))

	combined := append(append([]byte{}, certPEM...), keyPEM...)
	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = combined
	// r.TLS.Key is nil
	inst, err := r.New()
	assert.NoError(err)

	plan, err := inst.Plan(context.Background(), r)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
}

func Test_Plan_TLS_006(t *testing.T) {
	// Plan with cert+key concatenated in Key field, Cert nil - no error
	assert := assert.New(t)
	now := time.Now()
	certPEM, keyPEM := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))

	combined := append(append([]byte{}, certPEM...), keyPEM...)
	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	// r.TLS.Cert is nil
	r.TLS.Key = combined
	inst, err := r.New()
	assert.NoError(err)

	plan, err := inst.Plan(context.Background(), r)
	assert.NoError(err)
	assert.Equal(schema.ActionCreate, plan.Action)
}

func Test_Plan_TLS_007(t *testing.T) {
	// Plan with only cert (no key anywhere) - error
	assert := assert.New(t)
	now := time.Now()
	certPEM, _ := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))

	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	r.TLS.Cert = certPEM
	// r.TLS.Key is nil — no key anywhere, readPemBlocks will fail
	inst, err := r.New()
	assert.NoError(err)

	_, err = inst.Plan(context.Background(), r)
	assert.Error(err)
	assert.Contains(err.Error(), "tls")
}

func Test_Plan_TLS_008(t *testing.T) {
	// Plan with only key (no cert anywhere) - error
	assert := assert.New(t)
	now := time.Now()
	_, keyPEM := generateTestCert(t, now.Add(-1*time.Hour), now.Add(24*time.Hour))

	r := &resource.Resource{
		Listen: "localhost:8080",
		Router: &mockRouter{name: "r1", resourceName: "httprouter"},
	}
	// r.TLS.Cert is nil — no cert anywhere
	r.TLS.Key = keyPEM
	inst, err := r.New()
	assert.NoError(err)

	_, err = inst.Plan(context.Background(), r)
	assert.Error(err)
	assert.Contains(err.Error(), "tls")
}
