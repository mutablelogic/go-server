package nginx

import (
	"context"
	"net/http"
	"regexp"

	// Module imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (nginx *t) RegisterHandlers(parent context.Context, router plugin.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Get nginx status (version, uptime, available and enabled configurations)
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Return nginx status"),
		reRoot, nginx.ReqHealth,
		http.MethodGet,
	)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Label returns the label of nginx
func (nginx *t) Label() string {
	return nginx.label
}

// Description returns the label of the router
func (nginx *t) Description() string {
	return "Starts, stops and configures nginx reverse proxy"
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (nginx *t) ReqHealth(w http.ResponseWriter, r *http.Request) {
	util.ServeJSON(w, "ok", http.StatusOK, 2)
}
