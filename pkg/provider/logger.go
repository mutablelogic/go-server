package provider

import (
	"context"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Logger interface {
	hcl.Resource

	// Print logging message
	Print(context.Context, ...any)

	// Print logging message with format
	Printf(context.Context, string, ...any)
}
