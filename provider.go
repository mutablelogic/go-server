package server

import (
	"context"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
)

// Logger
type Logger interface {
	hcl.Resource

	// Print logging message
	Print(context.Context, ...any)

	// Print logging message with format
	Printf(context.Context, string, ...any)
}

// Provider
type Provider interface {
	hcl.Resource
	Logger

	// Set an attribute on a block
	Set(hcl.Block, hcl.Label, any) error
}
