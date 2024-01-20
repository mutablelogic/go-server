package hcl

import "context"

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Resource represents a managed resource
type Resource interface {
	Run(context.Context) error
}
