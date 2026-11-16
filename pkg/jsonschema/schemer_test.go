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

// channel has a field which can't be inferred, so it needs its own schema
type channel struct {
	C chan int
}

func (channel) JSONSchema() *jsonschema.Schema {
	return mustSchema(`{"type":"string"}`)
}

// node is recursive, which can't be inferred, so it needs its own schema
type node struct {
	Children []node
}

func (node) JSONSchema() *jsonschema.Schema {
	return mustSchema(`{"type":"array"}`)
}

// referenced has local references, which wouldn't be where it's used
type referenced struct{}

func (referenced) JSONSchema() *jsonschema.Schema {
	return mustSchema(`{"$defs":{"v":{"type":"string"}},"$ref":"#/$defs/v"}`)
}

// level returns the same schema each time, with values which could be
// shared
type level int

var levelSchema = mustSchema(`{"type":"object","properties":{"value":{"type":"integer","minimum":1,"enum":[1,2],"examples":[1]}}}`)

func (level) JSONSchema() *jsonschema.Schema {
	return levelSchema
}

func Test_SchemerReview(t *testing.T) {
	t.Run("NotInferred", func(t *testing.T) {
		// Its fields aren't looked into, so they needn't be inferable
		type withChannel struct {
			Channel channel `json:"channel"`
			Tree    []node  `json:"tree"`
		}
		schema, err := jsonschema.For[withChannel]()
		require.NoError(t, err)
		assert.NoError(t, schema.Validate(json.RawMessage(`{"channel":"a","tree":[[]]}`)))
		assert.Error(t, schema.Validate(json.RawMessage(`{"channel":1,"tree":[]}`)))
		_, err = jsonschema.For[node]()
		assert.NoError(t, err, "recursive")
	})

	t.Run("LocalReferences", func(t *testing.T) {
		type withReferenced struct {
			Value referenced `json:"value"`
		}
		_, err := jsonschema.For[withReferenced]()
		assert.ErrorContains(t, err, "local references")
		_, err = jsonschema.For[referenced]()
		assert.ErrorContains(t, err, "local references")
	})

	t.Run("Description", func(t *testing.T) {
		// A field's jsonschema tag is its description, before help
		type described struct {
			Attrs values `json:"attrs" jsonschema:"From the jsonschema tag" help:"From help"`
		}
		schema := jsonschema.MustFor[described]()
		assert.Equal(t, "From the jsonschema tag", schema.Properties["attrs"].Description)
		require.NotNil(t, schema.Properties["attrs"].AdditionalProperties, "its own schema")
	})

	t.Run("Shared", func(t *testing.T) {
		// Each use is a deep copy, so a change to one's values isn't the
		// provider's, or another use's
		type twice struct {
			A level `json:"a"`
			B level `json:"b"`
		}
		schema := jsonschema.MustFor[twice]()
		a := schema.Properties["a"].Properties["value"]
		b := schema.Properties["b"].Properties["value"]
		require.NotNil(t, a.Minimum)
		*a.Minimum = 99
		a.Enum[0] = 99
		a.Examples[0] = 99
		for name, s := range map[string]*jsonschema.Schema{"b": {Schema: *b}, "provider": {Schema: *levelSchema.Properties["value"]}} {
			assert.Equal(t, float64(1), *s.Minimum, name)
			assert.EqualValues(t, 1, s.Enum[0], name)
			assert.EqualValues(t, 1, s.Examples[0], name)
		}
	})
}
