package tokenauth

import (
	"context"
	"net/http"
	"regexp"

	// Module imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"
)

// Check to make sure interface is implemented for gatewat
var _ plugin.Gateway = (*tokenauth)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot = regexp.MustCompile(`^/?$`)
	reName = regexp.MustCompile(`^/` + types.ReIdentifier + `/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for tokenauth
func (tokenauth *tokenauth) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   List all tokens. Requires read scope.
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "List all tokens"),
		reRoot, tokenauth.ReqListAll,
		http.MethodGet,
	)

	// GET /:name:
	//   List token by name. Requires read scope.
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Retrive a token by name"),
		reName, tokenauth.ReqList,
		http.MethodGet,
	)

	// POST /:name:
	//   Create a new token, body can contain 'duration' and 'scopes', returns
	//   value. Requires write scope.
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeWrite), "Create a new token and return the token value"),
		reName, tokenauth.ReqCreate,
		http.MethodPost,
	)

	// PATCH /:name:
	//   Update a new token, body can contain 'duration' and 'scopes', returns
	//   token. Requires write scope.
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeWrite), "Update token duration or authorization scopes"),
		reName, tokenauth.ReqUpdate,
		http.MethodPatch,
	)

	// DELETE /:name:
	//   Delete a token, or rotate if admin token. Requires write scope.
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, ScopeWrite), "Delete a token by name, or rotate value if admin token"),
		reName, tokenauth.ReqDelete,
		http.MethodDelete,
	)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return description of the gateway service
func (tokenauth *tokenauth) Description() string {
	return "HTTP server token authentication"
}

// Return label of the gateway service
func (tokenauth *tokenauth) Label() string {
	return tokenauth.label
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (tokenauth *tokenauth) ReqListAll(w http.ResponseWriter, r *http.Request) {
	util.ServeError(w, http.StatusNotImplemented)
}

func (tokenauth *tokenauth) ReqList(w http.ResponseWriter, r *http.Request) {
	util.ServeError(w, http.StatusNotImplemented)
}

func (tokenauth *tokenauth) ReqCreate(w http.ResponseWriter, r *http.Request) {
	util.ServeError(w, http.StatusNotImplemented)
}

func (tokenauth *tokenauth) ReqUpdate(w http.ResponseWriter, r *http.Request) {
	util.ServeError(w, http.StatusNotImplemented)
}

func (tokenauth *tokenauth) ReqDelete(w http.ResponseWriter, r *http.Request) {
	util.ServeError(w, http.StatusNotImplemented)
}
