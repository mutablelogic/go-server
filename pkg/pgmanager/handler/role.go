package handler

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterRole(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgmanager.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "role"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodGet:
			_ = roleList(w, r, manager)
		case http.MethodPost:
			_ = roleCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "role/{name}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		// Parse request argument
		name := r.PathValue("name")
		if name == "" || !types.IsIdentifier(name) {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing or invalid role name"), name)
			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = roleGet(w, r, manager, name)
		case http.MethodDelete:
			_ = roleDelete(w, r, manager, name)
		case http.MethodPatch:
			_ = roleUpdate(w, r, manager, name)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func roleGet(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	role, err := manager.GetRole(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), role)
}

func roleDelete(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	_, err := manager.DeleteRole(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

func roleUpdate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	// Parse request
	var req schema.RoleMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Perform update
	role, err := manager.UpdateRole(r.Context(), name, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), role)
}

func roleCreate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req schema.RoleMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the role
	response, err := manager.CreateRole(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func roleList(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req schema.RoleListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the roles
	response, err := manager.ListRoles(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
