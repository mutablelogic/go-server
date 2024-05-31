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
	reName = regexp.MustCompile(`^/(` + types.ReIdentifier + `)/?$`)
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
		ctx.WithDescription(ctx.WithScope(parent, ScopeRead), "Retreive a token by name"),
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

// Serve all tokens (excluding the token value)
func (tokenauth *tokenauth) ReqListAll(w http.ResponseWriter, r *http.Request) {
	result := make([]Token, 0, len(tokenauth.tokens))
	for key, token := range tokenauth.tokens {
		result = append(result, Token{Name: key, Scope: token.Scope, Time: token.Time, Expire: token.Expire})
	}
	util.ServeJSON(w, result, http.StatusOK, 2)
}

// Serve one token. Returns 404 if not found
func (tokenauth *tokenauth) ReqList(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Retrieve token
	token, ok := tokenauth.tokens[params[0]]
	if !ok {
		util.ServeError(w, http.StatusNotFound)
		return
	}

	// Serve token
	util.ServeJSON(w, &Token{Name: params[0], Scope: token.Scope, Time: token.Time, Expire: token.Expire}, http.StatusOK, 2)
}

func (tokenauth *tokenauth) ReqCreate(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	// If the token already exists, return 409
	if _, ok := tokenauth.tokens[params[0]]; ok {
		util.ServeError(w, http.StatusConflict, "Token already exists")
		return
	}

	// Read in the body if there is one - require Duration and Scopes
	// then create the token and return the value
	var body TokenCreate
	if _, err := util.ReqDecodeBody(r, &body); err != nil {
		util.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else if value, err := tokenauth.Create(params[0], body.Duration, body.Scope...); err != nil {
		util.ServeError(w, http.StatusBadRequest, "Invalid duration or scopes")
		return
	} else {
		util.ServeText(w, value, http.StatusCreated)
	}
}

func (tokenauth *tokenauth) ReqUpdate(w http.ResponseWriter, r *http.Request) {
	// TODO
	util.ServeError(w, http.StatusNotImplemented)
}

func (tokenauth *tokenauth) ReqDelete(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Validate the token
	_, ok := tokenauth.tokens[params[0]]
	if !ok {
		util.ServeError(w, http.StatusNotFound)
		return
	}

	// Revoke the token
	if err := tokenauth.Revoke(params[0]); err != nil {
		util.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 202
	util.ServeError(w, http.StatusAccepted)
}
