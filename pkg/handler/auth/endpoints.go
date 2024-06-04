package auth

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
func (service *auth) AddEndpoints(ctx context.Context, router server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Get current set of tokens and groups
	router.AddHandlerFuncRe(ctx, reRoot, service.GetTokens, http.MethodGet)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get all tokens
func (service *auth) GetTokens(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, service.jar.Tokens(), http.StatusOK, jsonIndent)
}
