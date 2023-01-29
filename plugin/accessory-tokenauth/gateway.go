package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	// Module imports
	auth "github.com/mutablelogic/go-accessory/pkg/auth"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-accessory"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TokenUpdate struct {
	Duration time.Duration `json:"duration,omitempty"`
	Scope    []string      `json:"scope,omitempty"`
}

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
		ctx.WithDescription(ctx.WithScope(parent, auth.ScopeRead), "List all tokens"),
		reRoot, tokenauth.ReqListAll,
		http.MethodGet,
	)
	// POST /:name
	//   Create a new token and return it
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, auth.ScopeWrite), "Create a token and return the value"),
		reName, tokenauth.ReqCreate,
		http.MethodPost,
	)
	// DELETE /:name
	//   Delete a token
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, auth.ScopeWrite), "Delete a token"),
		reName, tokenauth.ReqDelete,
		http.MethodDelete,
	)
	// PATCH /:name
	//   Patch a token with duration and scope
	router.AddHandler(
		ctx.WithDescription(ctx.WithScope(parent, auth.ScopeWrite), "Update a token duration or scope"),
		reName, tokenauth.ReqPatch,
		http.MethodPatch,
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
	var result []AuthToken

	// List all tokens
	if err := tokenauth.List(r.Context(), func(token AuthToken) {
		result = append(result, token)
	}); err != nil {
		util.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Serve tokens
	util.ServeJSON(w, result, http.StatusOK, 2)
}

// Create a new token, optionally with duration and scope and return the
// token value as text/plain
func (tokenauth *tokenauth) ReqCreate(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	// Decode the body if one is present
	var body TokenUpdate
	if r.Body != nil && r.ContentLength != 0 {
		if _, err := util.ReqDecodeBody(r, &body); err != nil {
			util.ServeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Create the token
	value, err := tokenauth.CreateByte16(r.Context(), params[0], body.Duration, body.Scope...)
	if err != nil {
		if errors.Is(err, ErrDuplicateEntry) {
			util.ServeError(w, http.StatusConflict, err.Error())
		} else {
			util.ServeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Serve the token value as text/plain
	util.ServeText(w, value, http.StatusCreated)
}

func (tokenauth *tokenauth) ReqDelete(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	// If the token is the admin token, then return 403
	if params[0] == defaultAdmin {
		util.ServeError(w, http.StatusForbidden)
		return
	}

	// Delete the token
	err := tokenauth.Expire(r.Context(), params[0], true)
	if errors.Is(err, ErrNotFound) {
		util.ServeError(w, http.StatusNotFound)
		return
	} else if err != nil {
		util.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 202
	util.ServeError(w, http.StatusAccepted)
}

func (tokenauth *tokenauth) ReqPatch(w http.ResponseWriter, r *http.Request) {
	// Check parameters
	_, _, params := util.ReqPrefixPathParams(r)
	if len(params) != 1 {
		util.ServeError(w, http.StatusInternalServerError)
		return
	}

	// A body is required
	if r.Body == nil || r.ContentLength == 0 {
		util.ServeError(w, http.StatusBadRequest)
		return
	}

	// Decode the body
	var body TokenUpdate
	patch, err := util.ReqDecodeBody(r, &body)
	if err != nil {
		util.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Println("patch", patch, body)

	// Return 202
	util.ServeError(w, http.StatusAccepted)
}
