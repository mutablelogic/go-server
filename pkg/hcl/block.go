package hcl

import "context"

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Block represents a configuration block which can be created into a running
// task after the configuration has been parsed, validated and references between
// blocks resolved.
type Block interface {
	// Return the name for the block
	Name() string

	// Return a description for the block
	Description() string

	// Create a resource from a block
	New(context.Context) (Resource, error)
}
