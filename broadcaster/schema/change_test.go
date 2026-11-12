package schema_test

import (
	"encoding/json"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_ChangeFilter(t *testing.T) {
	created := schema.Change{Action: schema.ChangeCreate, Schema: "llm", Table: "provider", Row: "ollama"}

	t.Run("Matches", func(t *testing.T) {
		// A filter matches the changes with its schema, table and action,
		// each of which matches any when it's empty
		assert.True(t, schema.ChangeFilter{}.Matches(created))
		assert.True(t, schema.ChangeFilter{Schema: "llm", Table: "provider", Action: schema.ChangeCreate}.Matches(created))
		assert.False(t, schema.ChangeFilter{Schema: "memory"}.Matches(created))
		assert.False(t, schema.ChangeFilter{Table: "session"}.Matches(created))
		assert.False(t, schema.ChangeFilter{Action: schema.ChangeDelete}.Matches(created))
	})

	t.Run("Query", func(t *testing.T) {
		assert.Empty(t, schema.ChangeFilter{}.Query())
		assert.Equal(t, "action=update&schema=llm&table=session", schema.ChangeFilter{Schema: "llm", Table: "session", Action: schema.ChangeUpdate}.Query().Encode())
	})

	t.Run("Validate", func(t *testing.T) {
		assert.NoError(t, schema.ChangeFilter{}.Validate())
		assert.NoError(t, schema.ChangeFilter{Action: schema.ChangeDelete}.Validate())
		assert.Error(t, schema.ChangeFilter{Action: "INSERT"}.Validate())
	})
}

func Test_Change_JSON(t *testing.T) {
	t.Run("Change", func(t *testing.T) {
		// A change is its action, and what changed, which is left out when
		// it isn't set
		for _, test := range []struct {
			change schema.Change
			json   string
		}{
			{schema.Change{Action: schema.ChangeCreate, Schema: "llm", Table: "provider", Row: "ollama"}, `{"action":"create","schema":"llm","table":"provider","row":"ollama"}`},
			{schema.Change{Action: schema.ChangeUpdate, Table: "model"}, `{"action":"update","table":"model"}`},
			{schema.Change{Action: schema.ChangeDelete}, `{"action":"delete"}`},
			{schema.Change{Action: schema.ChangeUpdate, Table: "session", Row: "1", Data: json.RawMessage(`{"serial":2}`)}, `{"action":"update","table":"session","row":"1","data":{"serial":2}}`},
		} {
			data, err := json.Marshal(test.change)
			require.NoError(t, err)
			assert.JSONEq(t, test.json, string(data))

			var out schema.Change
			require.NoError(t, json.Unmarshal(data, &out))
			assert.Equal(t, test.change, out)
		}
	})

	t.Run("Unknown", func(t *testing.T) {
		// An action must be one of the actions
		var out schema.Change
		assert.Error(t, json.Unmarshal([]byte(`{"action":"INSERT","table":"model"}`), &out))
		var action schema.ChangeAction
		assert.Error(t, json.Unmarshal([]byte(`"truncate"`), &action))
		assert.Error(t, json.Unmarshal([]byte(`1`), &action))
	})
}
