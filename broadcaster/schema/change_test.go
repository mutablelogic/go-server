package schema_test

import (
	"encoding/json"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

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
