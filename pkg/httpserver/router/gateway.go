package router

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
	reRoot = regexp.MustCompile(`^/$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for router
func (gateway *router) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   Return health status for the router
	router.AddHandler(ctx.WithScope(parent, ScopeRead), reRoot, gateway.ReqHealth, http.MethodGet)
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (gateway *router) ReqHealth(w http.ResponseWriter, r *http.Request) {
	util.ServeError(w, http.StatusNotImplemented)
}
