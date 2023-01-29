package log

import (
	"context"
	"net/http"
	"time"

	// Module imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check to make sure interface is implemented for gatewat
var _ plugin.Gateway = (*t)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for tokenauth
func (log *t) RegisterHandlers(parent context.Context, router plugin.Router) {
	// tokenauth middleware
	router.AddMiddleware(
		ctx.WithDescription(ctx.WithPrefix(parent, "log"), "Log request, status and duration"),
		log.Middleware,
	)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of the router
func (log *t) Label() string {
	return log.label
}

// Description returns the label of the router
func (log *t) Description() string {
	return "Provides logging services"
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

// Register middleware for tokenauth
func (log *t) Middleware(child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		// Continue with the child handler
		child(w, r)

		// Log the request
		log.Printf(r.Context(), "%v %v => %v", r.Method, r.URL, time.Since(now).Truncate(time.Millisecond))
	}
}
