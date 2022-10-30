package plugin

import "context"

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Log plugin
type Log interface {
	Print(context.Context, ...interface{})
}
