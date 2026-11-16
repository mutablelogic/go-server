package jsonschema_test

import (
	"encoding/json"
	"testing"

	// Packages
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// values is a map whose values are strings, or {"base64": "..."}, as its
// own MarshalJSON would write them, which its fields don't say
type values map[string][]string

func (values) JSONSchema() *jsonschema.Schema {
	return mustSchema(`{"type":"object","additionalProperties":{"type":"array","items":{"anyOf":[{"type":"string"},{"type":"object","properties":{"base64":{"type":"string"}},"required":["base64"],"additionalProperties":false}]}}}`)
}

// version is a struct which is a string, with a pointer receiver
type version struct {
	Major, Minor int
}

func (*version) JSONSchema() *jsonschema.Schema {
	return mustSchema(`{"type":"string","pattern":"^[0-9]+\\.[0-9]+$"}`)
}

// container has the types, wherever they're used
type container struct {
	Attrs   values            `json:"attrs" help:"The attributes"`
	List    []values          `json:"list,omitempty"`
	ByName  map[string]values `json:"by_name,omitempty"`
	Pointer *values           `json:"pointer,omitempty"`
	Version version           `json:"version"`
	Plain   map[string]int    `json:"plain,omitempty"`
}

func mustSchema(data string) *jsonschema.Schema {
	schema, err := jsonschema.FromJSON(json.RawMessage(data))
	if err != nil {
		panic(err)
	}
	return schema
}

func Test_Schemer(t *testing.T) {
	t.Run("TopLevel", func(t *testing.T) {
		schema := jsonschema.MustFor[values]()
		assert.NoError(t, schema.Validate(json.RawMessage(`{"cn":["a",{"base64":"AQ=="}]}`)))
		assert.Error(t, schema.Validate(json.RawMessage(`{"cn":[1]}`)))
		assert.Error(t, schema.Validate(json.RawMessage(`{"cn":[{"hex":"01"}]}`)))

		// With a pointer receiver
		schema = jsonschema.MustFor[version]()
		assert.NoError(t, schema.Validate(json.RawMessage(`"1.2"`)))
		assert.Error(t, schema.Validate(json.RawMessage(`{"Major":1,"Minor":2}`)))
	})

	t.Run("Anywhere", func(t *testing.T) {
		// As a field, in a slice, in a map, and by a pointer
		schema := jsonschema.MustFor[container]()
		assert.NoError(t, schema.Validate(json.RawMessage(`{
			"attrs": {"cn": ["a", {"base64": "AQ=="}]},
			"list": [{"cn": [{"base64": "AQ=="}]}],
			"by_name": {"x": {"cn": ["b"]}},
			"pointer": {"cn": ["c"]},
			"version": "1.2",
			"plain": {"a": 1}
		}`)))
		for name, data := range map[string]string{
			"field":   `{"attrs":{"cn":[1]},"version":"1.2"}`,
			"slice":   `{"attrs":{},"list":[{"cn":[1]}],"version":"1.2"}`,
			"map":     `{"attrs":{},"by_name":{"x":{"cn":[1]}},"version":"1.2"}`,
			"pointer": `{"attrs":{},"pointer":{"cn":[1]},"version":"1.2"}`,
			"struct":  `{"attrs":{},"version":{"Major":1}}`,
			"others":  `{"attrs":{},"version":"1.2","plain":{"a":"x"}}`,
		} {
			assert.Error(t, schema.Validate(json.RawMessage(data)), name)
		}

		// A field of the type has its own tags
		assert.Equal(t, "The attributes", schema.Properties["attrs"].Description)
	})

	t.Run("Copied", func(t *testing.T) {
		// Each use has its own copy, so a change to one isn't another's
		a := jsonschema.MustFor[container]()
		a.Properties["attrs"].Description = "changed"
		b := jsonschema.MustFor[values]()
		assert.Empty(t, b.Description)
		require.NotNil(t, b.AdditionalProperties)
	})
}
