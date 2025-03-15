package provider

import (
	"context"
	"log"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Emit an informational message
func (provider *provider) Print(ctx context.Context, a ...any) {
	log.Print(a...)
}
