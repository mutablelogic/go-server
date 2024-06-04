package auth

import (
	"context"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.Middleware = (*auth)(nil)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (middleware *auth) Wrap(ctx context.Context, next http.HandlerFunc) http.HandlerFunc {
	logger := provider.Logger(ctx)
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf(ctx, "TODO Auth: %v", r.URL)
		next(w, r)
	}
}
