package schema_test

import (
	"encoding/json"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/openapi/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_NewSpec_001(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("My API", "2.0.0")
	assert.NotNil(spec)
	assert.Equal(schema.Version, spec.Openapi)
	assert.Equal("My API", spec.Info.Title)
	assert.Equal("2.0.0", spec.Info.Version)
	assert.Nil(spec.Paths)
	assert.Empty(spec.Servers)
}

func Test_AddPath_001(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.AddPath("/health", &schema.PathItem{
		Get: &schema.Operation{Summary: "Health check"},
	})
	assert.NotNil(spec.Paths)
	assert.Len(spec.Paths.MapOfPathItemValues, 1)
	item, ok := spec.Paths.MapOfPathItemValues["/health"]
	assert.True(ok)
	assert.NotNil(item.Get)
	assert.Equal("Health check", item.Get.Summary)
}

func Test_AddPath_002(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.AddPath("/nothing", nil)
	assert.NotNil(spec.Paths)
	assert.Empty(spec.Paths.MapOfPathItemValues)
}

func Test_AddPath_003(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.AddPath("/a", &schema.PathItem{
		Get: &schema.Operation{Summary: "A"},
	})
	spec.AddPath("/b", &schema.PathItem{
		Post: &schema.Operation{Summary: "B"},
	})
	assert.Len(spec.Paths.MapOfPathItemValues, 2)
}

func Test_AddPath_004(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.AddPath("/x", &schema.PathItem{
		Get: &schema.Operation{Summary: "first"},
	})
	spec.AddPath("/x", &schema.PathItem{
		Get: &schema.Operation{Summary: "second"},
	})
	assert.Len(spec.Paths.MapOfPathItemValues, 1)
	assert.Equal("second", spec.Paths.MapOfPathItemValues["/x"].Get.Summary)
}

func Test_SetServers_001(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.SetServers([]schema.Server{
		{URL: "http://localhost:8080"},
		{URL: "https://api.example.com", Description: "Production"},
	})
	assert.Len(spec.Servers, 2)
	assert.Equal("http://localhost:8080", spec.Servers[0].URL)
	assert.Equal("https://api.example.com", spec.Servers[1].URL)
	assert.Equal("Production", spec.Servers[1].Description)
}

func Test_SetServers_002(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.SetServers([]schema.Server{{URL: "http://old"}})
	spec.SetServers([]schema.Server{{URL: "http://new"}})
	assert.Len(spec.Servers, 1)
	assert.Equal("http://new", spec.Servers[0].URL)
}

func Test_SetServers_003(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("test", "1.0")
	spec.SetServers([]schema.Server{{URL: "http://x"}})
	spec.SetServers(nil)
	assert.Empty(spec.Servers)
}

func Test_Paths_MarshalJSON_001(t *testing.T) {
	assert := assert.New(t)

	paths := schema.Paths{
		MapOfPathItemValues: map[string]schema.PathItem{
			"/health": {Get: &schema.Operation{Summary: "Health"}},
		},
	}
	data, err := json.Marshal(paths)
	assert.NoError(err)

	var raw map[string]json.RawMessage
	assert.NoError(json.Unmarshal(data, &raw))
	_, hasHealth := raw["/health"]
	assert.True(hasHealth, "expected /health key in JSON output")
	_, hasMap := raw["MapOfPathItemValues"]
	assert.False(hasMap, "MapOfPathItemValues should not appear in JSON")
}

func Test_Paths_UnmarshalJSON_001(t *testing.T) {
	assert := assert.New(t)

	input := `{"/users":{"get":{"summary":"List users"}},"/items":{"post":{"summary":"Create item"}}}`
	var paths schema.Paths
	err := json.Unmarshal([]byte(input), &paths)
	assert.NoError(err)
	assert.Len(paths.MapOfPathItemValues, 2)

	users, ok := paths.MapOfPathItemValues["/users"]
	assert.True(ok)
	assert.NotNil(users.Get)
	assert.Equal("List users", users.Get.Summary)

	items, ok := paths.MapOfPathItemValues["/items"]
	assert.True(ok)
	assert.NotNil(items.Post)
	assert.Equal("Create item", items.Post.Summary)
}

func Test_Spec_RoundTrip_001(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("Round Trip", "3.0.0")
	spec.SetServers([]schema.Server{{URL: "http://localhost"}})
	spec.AddPath("/ping", &schema.PathItem{
		Get: &schema.Operation{
			Tags:    []string{"health"},
			Summary: "Ping",
		},
	})

	data, err := json.Marshal(spec)
	assert.NoError(err)

	var decoded schema.Spec
	assert.NoError(json.Unmarshal(data, &decoded))

	assert.Equal(spec.Openapi, decoded.Openapi)
	assert.Equal(spec.Info.Title, decoded.Info.Title)
	assert.Equal(spec.Info.Version, decoded.Info.Version)
	assert.Len(decoded.Servers, 1)
	assert.Equal("http://localhost", decoded.Servers[0].URL)
	assert.NotNil(decoded.Paths)
	assert.Len(decoded.Paths.MapOfPathItemValues, 1)

	ping := decoded.Paths.MapOfPathItemValues["/ping"]
	assert.NotNil(ping.Get)
	assert.Equal("Ping", ping.Get.Summary)
	assert.Equal([]string{"health"}, ping.Get.Tags)
}

func Test_Spec_RoundTrip_002(t *testing.T) {
	assert := assert.New(t)

	spec := schema.NewSpec("Empty", "0.1")
	data, err := json.Marshal(spec)
	assert.NoError(err)

	var decoded schema.Spec
	assert.NoError(json.Unmarshal(data, &decoded))
	assert.Equal("Empty", decoded.Info.Title)
	assert.Equal("0.1", decoded.Info.Version)
	assert.Nil(decoded.Paths)
	assert.Empty(decoded.Servers)
}

func Test_PathItem_AllMethods(t *testing.T) {
	assert := assert.New(t)

	item := schema.PathItem{
		Summary:     "All methods",
		Description: "Test all ops",
		Get:         &schema.Operation{Summary: "get"},
		Put:         &schema.Operation{Summary: "put"},
		Post:        &schema.Operation{Summary: "post"},
		Delete:      &schema.Operation{Summary: "delete"},
		Options:     &schema.Operation{Summary: "options"},
		Head:        &schema.Operation{Summary: "head"},
		Patch:       &schema.Operation{Summary: "patch"},
		Trace:       &schema.Operation{Summary: "trace"},
	}

	data, err := json.Marshal(item)
	assert.NoError(err)

	var decoded schema.PathItem
	assert.NoError(json.Unmarshal(data, &decoded))
	assert.Equal("get", decoded.Get.Summary)
	assert.Equal("put", decoded.Put.Summary)
	assert.Equal("post", decoded.Post.Summary)
	assert.Equal("delete", decoded.Delete.Summary)
	assert.Equal("options", decoded.Options.Summary)
	assert.Equal("head", decoded.Head.Summary)
	assert.Equal("patch", decoded.Patch.Summary)
	assert.Equal("trace", decoded.Trace.Summary)
}
