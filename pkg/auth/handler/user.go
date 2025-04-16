package handler

import (
	"context"
	"net/http"

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

func RegisterUser(ctx context.Context, router server.HTTPRouter, prefix string, manager *auth.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "user"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = userList(w, r, manager)
		case http.MethodPost:
			_ = userCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "user/{name...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		// Get the name
		name := r.PathValue("name")
		if name == "" {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusBadRequest), "missing name")
			return
		}

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = userGet(w, r, manager, name)
		case http.MethodDelete:
			_ = userDelete(w, r, manager, name)
		case http.MethodPatch:
			_ = userUpdate(w, r, manager, name)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// List all users
func userList(w http.ResponseWriter, r *http.Request, manager *auth.Manager) error {
	// Parse request
	var req schema.UserListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the users
	response, err := manager.ListUsers(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

// Create a new user
func userCreate(w http.ResponseWriter, r *http.Request, manager *auth.Manager) error {
	// Parse request
	var req schema.UserMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the user
	response, err := manager.CreateUser(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

// Get a user
func userGet(w http.ResponseWriter, r *http.Request, manager *auth.Manager, name string) error {
	// Get the user
	response, err := manager.GetUser(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

// Delete a user
func userDelete(w http.ResponseWriter, r *http.Request, manager *auth.Manager, name string) error {
	// Parse request
	var req struct {
		Force bool `json:"force"`
	}
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Delete the user
	_, err := manager.DeleteUser(r.Context(), name, req.Force)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

// Update a user
func userUpdate(w http.ResponseWriter, r *http.Request, manager *auth.Manager, name string) error {
	// Parse request
	var req schema.UserMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Update the user
	response, err := manager.UpdateUser(r.Context(), name, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
