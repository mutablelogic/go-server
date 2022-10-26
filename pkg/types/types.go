package types

import (
	"time"

	// Package imports
	iface "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Bool bool
type Int int64
type UInt uint64
type Float float64
type Duration time.Duration
type String string

// Task is a Task type, which can be either a reference to a task by name, or
// the instance. Binding from reference to an instance is done after configuration
// is parsed.
type Task struct {
	iface.Task
	Ref string
}
