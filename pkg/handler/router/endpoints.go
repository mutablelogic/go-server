package router

import (
	"context"
	"net/http"
	"regexp"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reRoot = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *router) AddEndpoints(ctx context.Context, r server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Get router services
	r.AddHandlerFuncRe(ctx, reRoot, service.GetScopes, http.MethodGet).(Route).
		SetScope(service.ScopeRead()...)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get registered scopes
func (service *router) GetScopes(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, service.Scopes(), http.StatusOK, jsonIndent)
}
