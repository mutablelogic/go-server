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

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reRoot = regexp.MustCompile(`^/$`)
	reName = regexp.MustCompile(`^/` + types.ReIdentifier + `/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Register routes for tokenauth
func (tokenauth *tokenauth) RegisterHandlers(parent context.Context, router plugin.Router) {
	// GET /
	//   List all tokens. Requires read scope.
	router.AddHandler(ctx.WithScope(parent, ScopeRead), reRoot, tokenauth.ReqListAll, http.MethodGet)

	// GET /:name:
	//   List token by name. Requires read scope.
	router.AddHandler(ctx.WithScope(parent, ScopeRead), reName, tokenauth.ReqList, http.MethodGet)

	// POST /:name:
	//   Create a new token, body can contain 'duration' and 'scopes', returns
	//   value. Requires write scope.
	router.AddHandler(ctx.WithScope(parent, ScopeWrite), reName, tokenauth.ReqCreate, http.MethodPost)

	// PATCH /:name:
	//   Update a new token, body can contain 'duration' and 'scopes', returns
	//   token. Requires write scope.
	router.AddHandler(ctx.WithScope(parent, ScopeWrite), reName, tokenauth.ReqUpdate, http.MethodPatch)

	// DELETE /:name:
	//   Delete a token, or rotate if admin token. Requires write scope.
	router.AddHandler(ctx.WithScope(parent, ScopeWrite), reName, tokenauth.ReqDelete, http.MethodPatch)
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
