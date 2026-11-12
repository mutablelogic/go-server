package cmd_test

import (
	"testing"

	// Packages
	kong "github.com/alecthomas/kong"
	cmd "github.com/mutablelogic/go-server/broadcaster/cmd"
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// The changes command parses its filter, and rejects an unknown action
func Test_ChangesCmd(t *testing.T) {
	var cli cmd.ClientCommands
	parser, err := kong.New(&cli)
	require.NoError(t, err)

	_, err = parser.Parse([]string{"changes", "--schema", "llm", "--table", "session", "--action", "update"})
	require.NoError(t, err)
	assert.Equal(t, schema.ChangeFilter{Schema: "llm", Table: "session", Action: schema.ChangeUpdate}, cli.Changes.ChangeFilter)

	_, err = parser.Parse([]string{"changes", "--action", "INSERT"})
	assert.Error(t, err)
}
