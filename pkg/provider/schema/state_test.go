package schema_test

import (
	"testing"
	"time"

	"github.com/mutablelogic/go-server/pkg/provider/schema"
	"github.com/mutablelogic/go-server/pkg/provider/schema/schematest"
	"github.com/stretchr/testify/assert"
)

func Test_Decode_001(t *testing.T) {
	assert := assert.New(t)
	// Non-pointer returns an error
	err := schema.State{}.Decode(struct{}{}, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "expected pointer to struct")
}

func Test_Decode_002(t *testing.T) {
	assert := assert.New(t)
	// Pointer to non-struct returns an error
	s := "hello"
	err := schema.State{}.Decode(&s, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "expected pointer to struct")
}

func Test_Decode_003(t *testing.T) {
	assert := assert.New(t)
	// Empty state into empty struct — no error
	var v struct{}
	assert.NoError(schema.State{}.Decode(&v, nil))
}

func Test_Decode_004(t *testing.T) {
	assert := assert.New(t)
	// Simple string field
	type simple struct {
		Name string `name:"name"`
	}
	state := schema.State{"name": "hello"}
	var v simple
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("hello", v.Name)
}

func Test_Decode_005(t *testing.T) {
	assert := assert.New(t)
	// Missing key leaves field at zero value
	type simple struct {
		Name string `name:"name"`
		Port int    `name:"port"`
	}
	state := schema.State{"name": "hello"}
	var v simple
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("hello", v.Name)
	assert.Equal(0, v.Port)
}

func Test_Decode_006(t *testing.T) {
	assert := assert.New(t)
	// Duration from string
	type withDuration struct {
		Timeout time.Duration `name:"timeout"`
	}
	state := schema.State{"timeout": "5m0s"}
	var v withDuration
	assert.NoError(state.Decode(&v, nil))
	assert.Equal(5*time.Minute, v.Timeout)
}

func Test_Decode_007(t *testing.T) {
	assert := assert.New(t)
	// Bool field from bool value
	type withBool struct {
		Enabled bool `name:"enabled"`
	}
	state := schema.State{"enabled": true}
	var v withBool
	assert.NoError(state.Decode(&v, nil))
	assert.True(v.Enabled)
}

func Test_Decode_008(t *testing.T) {
	assert := assert.New(t)
	// Bool field from string value
	type withBool struct {
		Enabled bool `name:"enabled"`
	}
	state := schema.State{"enabled": "true"}
	var v withBool
	assert.NoError(state.Decode(&v, nil))
	assert.True(v.Enabled)
}

func Test_Decode_009(t *testing.T) {
	assert := assert.New(t)
	// Embedded struct with prefix
	type withEmbed struct {
		TLS struct {
			Name   string `name:"name"`
			Verify bool   `name:"verify"`
		} `embed:"" prefix:"tls."`
	}
	state := schema.State{
		"tls.name":   "example.com",
		"tls.verify": true,
	}
	var v withEmbed
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("example.com", v.TLS.Name)
	assert.True(v.TLS.Verify)
}

func Test_Decode_010(t *testing.T) {
	assert := assert.New(t)
	// Fields with name:"-" are skipped
	type withSkip struct {
		Hidden  string `name:"-"`
		Visible string `name:"visible"`
	}
	state := schema.State{"visible": "yes", "-": "no"}
	var v withSkip
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("yes", v.Visible)
	assert.Equal("", v.Hidden)
}

func Test_Decode_011(t *testing.T) {
	assert := assert.New(t)
	// Nil resolver with reference in state returns error
	r := httpserverResource{}
	state := schema.State{
		"listen":       ":8080",
		"router":       "router-01",
		"read-timeout": "1m0s",
	}
	err := state.Decode(&r, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "router")
	assert.Contains(err.Error(), "no resolver")
}

func Test_Decode_012(t *testing.T) {
	assert := assert.New(t)
	// float64 to int conversion (JSON numbers)
	type withInt struct {
		Port int `name:"port"`
	}
	state := schema.State{"port": float64(8080)}
	var v withInt
	assert.NoError(state.Decode(&v, nil))
	assert.Equal(8080, v.Port)
}

func Test_Decode_013(t *testing.T) {
	assert := assert.New(t)
	// String to int conversion
	type withInt struct {
		Port int `name:"port"`
	}
	state := schema.State{"port": "9090"}
	var v withInt
	assert.NoError(state.Decode(&v, nil))
	assert.Equal(9090, v.Port)
}

func Test_Decode_014(t *testing.T) {
	assert := assert.New(t)
	// Nil value in state leaves field untouched
	type simple struct {
		Name string `name:"name"`
	}
	v := simple{Name: "original"}
	state := schema.State{"name": nil}
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("original", v.Name)
}

func Test_Decode_015(t *testing.T) {
	assert := assert.New(t)
	// Full round-trip: StateOf then Decode (struct with no interface refs)
	type serverConfig struct {
		Listen       string        `name:"listen"`
		ReadTimeout  time.Duration `name:"read-timeout"`
		WriteTimeout time.Duration `name:"write-timeout"`
		TLS          struct {
			Name   string `name:"name"`
			Verify bool   `name:"verify"`
			Cert   string `name:"cert"`
			Key    string `name:"key"`
		} `embed:"" prefix:"tls."`
	}
	original := serverConfig{
		Listen:       "localhost:8080",
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Second,
	}
	original.TLS.Name = "myserver"
	original.TLS.Verify = true
	original.TLS.Cert = "/tmp/cert.pem"
	original.TLS.Key = "/tmp/key.pem"

	state := schema.StateOf(original)

	var decoded serverConfig
	assert.NoError(state.Decode(&decoded, nil))

	assert.Equal(original.Listen, decoded.Listen)
	assert.Equal(original.ReadTimeout, decoded.ReadTimeout)
	assert.Equal(original.WriteTimeout, decoded.WriteTimeout)
	assert.Equal(original.TLS.Name, decoded.TLS.Name)
	assert.Equal(original.TLS.Verify, decoded.TLS.Verify)
	assert.Equal(original.TLS.Cert, decoded.TLS.Cert)
	assert.Equal(original.TLS.Key, decoded.TLS.Key)
}

func Test_Decode_016(t *testing.T) {
	assert := assert.New(t)
	// Invalid duration string returns an error
	type withDuration struct {
		Timeout time.Duration `name:"timeout"`
	}
	state := schema.State{"timeout": "not-a-duration"}
	var v withDuration
	err := state.Decode(&v, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "timeout")
}

func Test_Decode_017(t *testing.T) {
	assert := assert.New(t)
	// Incompatible type returns an error
	type withString struct {
		Name string `name:"name"`
	}
	state := schema.State{"name": []int{1, 2, 3}}
	var v withString
	err := state.Decode(&v, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "name")
}

func Test_Decode_018(t *testing.T) {
	assert := assert.New(t)
	// String to float conversion
	type withFloat struct {
		Rate float64 `name:"rate"`
	}
	state := schema.State{"rate": "3.14"}
	var v withFloat
	assert.NoError(state.Decode(&v, nil))
	assert.InDelta(3.14, v.Rate, 0.001)
}

func Test_Decode_019(t *testing.T) {
	assert := assert.New(t)
	// Unexported fields are skipped
	type withUnexported struct {
		hidden  string `name:"hidden"`
		Visible string `name:"visible"`
	}
	state := schema.State{"visible": "yes", "hidden": "no"}
	var v withUnexported
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("yes", v.Visible)
	assert.Equal("", v.hidden)
}

func Test_Decode_020(t *testing.T) {
	assert := assert.New(t)
	// []any → []string (JSON round-trip)
	type withSlice struct {
		Tags []string `name:"tags"`
	}
	state := schema.State{"tags": []any{"a", "b", "c"}}
	var v withSlice
	assert.NoError(state.Decode(&v, nil))
	assert.Equal([]string{"a", "b", "c"}, v.Tags)
}

func Test_Decode_021(t *testing.T) {
	assert := assert.New(t)
	// []string → []string (direct assignment)
	type withSlice struct {
		Tags []string `name:"tags"`
	}
	state := schema.State{"tags": []string{"x", "y"}}
	var v withSlice
	assert.NoError(state.Decode(&v, nil))
	assert.Equal([]string{"x", "y"}, v.Tags)
}

func Test_Decode_022(t *testing.T) {
	assert := assert.New(t)
	// []any → []int (element-wise conversion from JSON floats)
	type withInts struct {
		Ports []int `name:"ports"`
	}
	state := schema.State{"ports": []any{float64(80), float64(443)}}
	var v withInts
	assert.NoError(state.Decode(&v, nil))
	assert.Equal([]int{80, 443}, v.Ports)
}

func Test_Decode_023(t *testing.T) {
	assert := assert.New(t)
	// map[string]any → map[string]any (direct assignment)
	type withMap struct {
		Meta map[string]any `name:"meta"`
	}
	state := schema.State{"meta": map[string]any{"key": "val", "n": float64(1)}}
	var v withMap
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("val", v.Meta["key"])
	assert.Equal(float64(1), v.Meta["n"])
}

func Test_Decode_024(t *testing.T) {
	assert := assert.New(t)
	// map[string]any → map[string]string (element-wise conversion)
	type withMap struct {
		Labels map[string]string `name:"labels"`
	}
	state := schema.State{"labels": map[string]any{"env": "prod", "app": "server"}}
	var v withMap
	assert.NoError(state.Decode(&v, nil))
	assert.Equal("prod", v.Labels["env"])
	assert.Equal("server", v.Labels["app"])
}

func Test_Decode_025(t *testing.T) {
	assert := assert.New(t)
	// Empty slice stays empty (not nil)
	type withSlice struct {
		Tags []string `name:"tags"`
	}
	state := schema.State{"tags": []any{}}
	var v withSlice
	assert.NoError(state.Decode(&v, nil))
	assert.NotNil(v.Tags)
	assert.Empty(v.Tags)
}

func Test_Decode_026(t *testing.T) {
	assert := assert.New(t)
	// Slice with incompatible element returns error
	type withSlice struct {
		Tags []string `name:"tags"`
	}
	state := schema.State{"tags": []any{"ok", []int{1}}}
	var v withSlice
	err := state.Decode(&v, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "tags")
}

func Test_Decode_027(t *testing.T) {
	assert := assert.New(t)
	// StateOf round-trip with slice
	type withSlice struct {
		Names []string `name:"names"`
	}
	original := withSlice{Names: []string{"alice", "bob"}}
	state := schema.StateOf(original)
	var decoded withSlice
	assert.NoError(state.Decode(&decoded, nil))
	assert.Equal(original.Names, decoded.Names)
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
	r.TLS.Cert = []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")
	r.TLS.Key = []byte("-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----")

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
	assert.Equal([]byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"), s["tls.cert"])
	assert.Equal([]byte("-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----"), s["tls.key"])
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
// Decode with Resolver tests

func Test_Decode_028(t *testing.T) {
	assert := assert.New(t)

	// Nil resolver with reference in state returns error
	state := schema.State{"listen": ":8080", "router": "router-01"}
	var v httpserverResource
	err := state.Decode(&v, nil)
	assert.Error(err)
	assert.Contains(err.Error(), "router")
	assert.Contains(err.Error(), "no resolver")
}

func Test_Decode_029(t *testing.T) {
	assert := assert.New(t)

	// Resolver resolves a reference by name
	mock := &schematest.ResourceInstance{N: "router-01", RN: "httprouter"}
	resolver := func(name string) schema.ResourceInstance {
		if name == "router-01" {
			return mock
		}
		return nil
	}

	state := schema.State{"listen": ":8080", "router": "router-01"}
	var v httpserverResource
	assert.NoError(state.Decode(&v, resolver))
	assert.Equal(":8080", v.Listen)
	assert.NotNil(v.Router)
	assert.Equal("router-01", v.Router.Name())
}

func Test_Decode_030(t *testing.T) {
	assert := assert.New(t)

	// Resolver returns nil for unknown name — required field returns error
	resolver := func(name string) schema.ResourceInstance {
		return nil
	}

	state := schema.State{"router": "nonexistent"}
	var v httpserverResource
	err := state.Decode(&v, resolver)
	assert.Error(err)
	assert.Contains(err.Error(), "router")
	assert.Contains(err.Error(), "not found")
}

func Test_Decode_031(t *testing.T) {
	assert := assert.New(t)

	// Required reference with empty/missing value in state — error
	resolver := func(name string) schema.ResourceInstance {
		return nil
	}

	state := schema.State{"listen": ":8080"} // no "router" key
	var v httpserverResource
	err := state.Decode(&v, resolver)
	// "router" is required:"" — missing from state with a resolver should error
	assert.Error(err)
	assert.Contains(err.Error(), "router")
	assert.Contains(err.Error(), "required")
}

func Test_Decode_032(t *testing.T) {
	assert := assert.New(t)

	// Optional interface field — resolver returns nil, no error
	type withOptionalRef struct {
		Listen string                  `name:"listen"`
		Logger schema.ResourceInstance `name:"logger" type:"log"`
	}
	resolver := func(name string) schema.ResourceInstance {
		return nil
	}

	state := schema.State{"listen": ":8080", "logger": "log-01"}
	var v withOptionalRef
	assert.NoError(state.Decode(&v, resolver))
	assert.Equal(":8080", v.Listen)
	assert.Nil(v.Logger)
}

func Test_Decode_033(t *testing.T) {
	assert := assert.New(t)

	// Round-trip: StateOf serialises reference name, Decode restores it
	mock := &schematest.ResourceInstance{N: "my-router", RN: "httprouter"}
	original := httpserverResource{
		Listen: ":443",
		Router: mock,
	}

	state := schema.StateOf(original)
	assert.Equal("my-router", state["router"]) // stored as name string

	resolver := func(name string) schema.ResourceInstance {
		if name == "my-router" {
			return mock
		}
		return nil
	}

	var decoded httpserverResource
	assert.NoError(state.Decode(&decoded, resolver))
	assert.Equal(":443", decoded.Listen)
	assert.NotNil(decoded.Router)
	assert.Equal("my-router", decoded.Router.Name())
}

///////////////////////////////////////////////////////////////////////////////
// Decode defaults tests

func Test_Decode_034(t *testing.T) {
	assert := assert.New(t)
	// Default applied when key is absent
	type withDefaults struct {
		Name string `name:"name" default:"world"`
	}
	var v withDefaults
	assert.NoError(schema.State{}.Decode(&v, nil))
	assert.Equal("world", v.Name)
}

func Test_Decode_035(t *testing.T) {
	assert := assert.New(t)
	// Default NOT applied when key is present
	type withDefaults struct {
		Name string `name:"name" default:"world"`
	}
	var v withDefaults
	assert.NoError(schema.State{"name": "hello"}.Decode(&v, nil))
	assert.Equal("hello", v.Name)
}

func Test_Decode_036(t *testing.T) {
	assert := assert.New(t)
	// Bool default from string tag
	type withDefaults struct {
		Enabled bool `name:"enabled" default:"true"`
	}
	var v withDefaults
	assert.NoError(schema.State{}.Decode(&v, nil))
	assert.True(v.Enabled)
}

func Test_Decode_037(t *testing.T) {
	assert := assert.New(t)
	// Bool default NOT applied when key is present (even if false)
	type withDefaults struct {
		Enabled bool `name:"enabled" default:"true"`
	}
	var v withDefaults
	assert.NoError(schema.State{"enabled": false}.Decode(&v, nil))
	assert.False(v.Enabled)
}

func Test_Decode_038(t *testing.T) {
	assert := assert.New(t)
	// Int default from string tag
	type withDefaults struct {
		Port int `name:"port" default:"8080"`
	}
	var v withDefaults
	assert.NoError(schema.State{}.Decode(&v, nil))
	assert.Equal(8080, v.Port)
}

func Test_Decode_039(t *testing.T) {
	assert := assert.New(t)
	// Duration default from string tag
	type withDefaults struct {
		Timeout time.Duration `name:"timeout" default:"5m"`
	}
	var v withDefaults
	assert.NoError(schema.State{}.Decode(&v, nil))
	assert.Equal(5*time.Minute, v.Timeout)
}

func Test_Decode_040(t *testing.T) {
	assert := assert.New(t)
	// Default in embedded struct with prefix
	type withEmbed struct {
		TLS struct {
			Verify bool `name:"verify" default:"true"`
		} `embed:"" prefix:"tls."`
	}
	var v withEmbed
	assert.NoError(schema.State{}.Decode(&v, nil))
	assert.True(v.TLS.Verify)
}

func Test_Decode_041(t *testing.T) {
	assert := assert.New(t)
	// Multiple defaults, only missing keys filled
	type withDefaults struct {
		Host string `name:"host" default:"localhost"`
		Port int    `name:"port" default:"443"`
	}
	var v withDefaults
	assert.NoError(schema.State{"host": "example.com"}.Decode(&v, nil))
	assert.Equal("example.com", v.Host)
	assert.Equal(443, v.Port)
}
