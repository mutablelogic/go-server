package types

import (
	"fmt"
	"strconv"

	// Package imports
	iface "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Task type, which can be either a reference to a task by name, or
// the instance. Binding from reference to an instance is done after configuration
// is parsed, in the provider.
type Task struct {
	iface.Task
	Ref string
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Task) String() string {
	if t.Task != nil {
		return fmt.Sprint(t.Task)
	} else if t.Ref != "" {
		return fmt.Sprintf("<task ref=%q>", t.Ref)
	} else {
		return "<nil>"
	}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *Task) UnmarshalJSON(data []byte) error {
	if v, err := strconv.Unquote(string(data)); err != nil {
		return err
	} else {
		t.Ref = v
		return nil
	}
}
