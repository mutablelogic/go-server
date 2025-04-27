package handler

import (
	"context"
	"net/http"
	"strconv"

	// Packages
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/auth"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterToken(ctx context.Context, router server.HTTPRouter, prefix string, manager *auth.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "token/{user}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		// Get the user
		user := r.PathValue("user")
		if user == "" {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusBadRequest), "missing user")
			return
		}

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = tokenList(w, r, manager, user)
		case http.MethodPost:
			_ = tokenCreate(w, r, manager, user)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "token/{user}/{id...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		// Get the user
		user := r.PathValue("user")
		if user == "" {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusBadRequest), "missing user")
			return
		}

		// Get the token ID
		id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusBadRequest), "invalid token ID")
			return
		}

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = tokenGet(w, r, manager, user, id)
		case http.MethodDelete:
			_ = tokenDelete(w, r, manager, user, id)
		case http.MethodPatch:
			_ = tokenUpdate(w, r, manager, user, id)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Get a token for a user
func tokenGet(w http.ResponseWriter, r *http.Request, manager *auth.Manager, user string, id uint64) error {
	// Get the token
	response, err := manager.GetToken(r.Context(), user, id)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

// Delete a token for a user
func tokenDelete(w http.ResponseWriter, r *http.Request, manager *auth.Manager, user string, id uint64) error {
	// Parse request
	var req struct {
		Force bool `json:"force"`
	}
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Delete the token
	_, err := manager.DeleteToken(r.Context(), user, id, req.Force)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

// Update a token for a user
func tokenUpdate(w http.ResponseWriter, r *http.Request, manager *auth.Manager, user string, id uint64) error {
	// Parse request
	var req schema.TokenMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Update the token
	response, err := manager.UpdateToken(r.Context(), user, id, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

// Create a token for a user
func tokenCreate(w http.ResponseWriter, r *http.Request, manager *auth.Manager, user string) error {
	// Parse request
	var req schema.TokenMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the token
	response, err := manager.CreateToken(r.Context(), user, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

// List tokens for a user
func tokenList(w http.ResponseWriter, r *http.Request, manager *auth.Manager, user string) error {
	// Parse request
	var req schema.TokenListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the tokens
	response, err := manager.ListTokens(r.Context(), user, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
