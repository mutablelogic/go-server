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

		switch r.Method {
		case http.MethodGet:
			_ = userList(w, r, manager)
		case http.MethodPost:
			_ = userCreate(w, r, manager)
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
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
