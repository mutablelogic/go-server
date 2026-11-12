// Package schema has the changes which a broadcaster publishes
package schema

import (
	"encoding/json"
	"fmt"
	"net/url"

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

// ChangeFilter matches changes: those of its schema, table and action, each
// of which matches any when it's empty
type ChangeFilter struct {
	Schema string       `json:"schema,omitempty" name:"schema" help:"Only the changes of the schema" optional:""`
	Table  string       `json:"table,omitempty" name:"table" help:"Only the changes of the table" optional:""`
	Action ChangeAction `json:"action,omitempty" name:"action" help:"Only the changes with the action: create, update or delete" optional:""`
}

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

func (f ChangeFilter) String() string {
	return types.Stringify(f)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Validate returns an error when the filter's action, when it has one,
// isn't one of the actions
func (f ChangeFilter) Validate() error {
	switch f.Action {
	case "", ChangeCreate, ChangeUpdate, ChangeDelete:
		return nil
	default:
		return fmt.Errorf("unknown change action: %q", f.Action)
	}
}

// Matches returns true when the change has the filter's schema, table and
// action, each of which matches any when it's empty
func (f ChangeFilter) Matches(change Change) bool {
	return (f.Schema == "" || f.Schema == change.Schema) &&
		(f.Table == "" || f.Table == change.Table) &&
		(f.Action == "" || f.Action == change.Action)
}

// Query returns the filter as the query of a request to watch changes
func (f ChangeFilter) Query() url.Values {
	values := url.Values{}
	if f.Schema != "" {
		values.Set("schema", f.Schema)
	}
	if f.Table != "" {
		values.Set("table", f.Table)
	}
	if f.Action != "" {
		values.Set("action", string(f.Action))
	}
	return values
}

////////////////////////////////////////////////////////////////////////////////
// JSONIFY

// UnmarshalJSON decodes an action, which must be one of the actions
func (a *ChangeAction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return a.UnmarshalText([]byte(s))
}

// UnmarshalText decodes an action, such as of a flag, or a query, which must
// be one of the actions
func (a *ChangeAction) UnmarshalText(data []byte) error {
	switch action := ChangeAction(data); action {
	case ChangeCreate, ChangeUpdate, ChangeDelete:
		*a = action
	default:
		return fmt.Errorf("unknown change action: %q", string(data))
	}
	return nil
}
