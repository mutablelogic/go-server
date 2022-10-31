package plugin

import "context"

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Log plugin
type Log interface {
	// Print log message
	Print(context.Context, ...interface{})

	// Format and print log message
	Printf(context.Context, string, ...interface{})
}
