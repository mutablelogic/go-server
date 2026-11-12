// Package schema has the changes which a broadcaster publishes
package schema

import (
	"encoding/json"
	"fmt"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Change is how something changed, which is its action, what changed, which
// is a table of a schema, and which row, or rows, changed, when it says,
// such as by an identifier which the rows share, and what the table says of
// the change, such as a row's new serial, when it says anything
type Change struct {
	Action ChangeAction    `json:"action" help:"How it changed, which is create, update or delete" example:"create"`
	Schema string          `json:"schema,omitempty" help:"Schema of the table which changed" example:"llm"`
	Table  string          `json:"table,omitempty" help:"Table which changed" example:"provider"`
	Row    string          `json:"row,omitempty" help:"Row, or rows, which changed, such as by an identifier" example:"ollama"`
	Data   json.RawMessage `json:"data,omitempty" help:"What the table says of the change, such as a row's new serial"`
}

// ChangeAction is how something changed: it was created, updated or deleted
type ChangeAction string

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// The events of a stream of a broadcaster's changes: each change, and that
// the server is stopping, before the stream ends, so a client knows it
// ended as there will be no more changes
const (
	ChangeName     = "change"
	ChangeShutdown = "shutdown"
)

const (
	ChangeCreate ChangeAction = "create" // Created
	ChangeUpdate ChangeAction = "update" // Updated
	ChangeDelete ChangeAction = "delete" // Deleted
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Change) String() string {
	return types.Stringify(c)
}

func (a ChangeAction) String() string {
	return string(a)
}

////////////////////////////////////////////////////////////////////////////////
// JSONIFY

// UnmarshalJSON decodes an action, which must be one of the actions
func (a *ChangeAction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch action := ChangeAction(s); action {
	case ChangeCreate, ChangeUpdate, ChangeDelete:
		*a = action
	default:
		return fmt.Errorf("unknown change action: %q", s)
	}
	return nil
}
