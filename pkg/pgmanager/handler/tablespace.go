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

func RegisterTablespace(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgmanager.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "tablespace"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodGet:
			_ = tablespaceList(w, r, manager)
		case http.MethodPost:
			_ = tablespaceCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "tablespace/{name}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		switch r.Method {
		case http.MethodGet:
			_ = tablespaceGet(w, r, manager, r.PathValue("name"))
		case http.MethodDelete:
			_ = tablespaceDelete(w, r, manager, r.PathValue("name"))
		case http.MethodPatch:
			_ = tablespaceUpdate(w, r, manager, r.PathValue("name"))
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func tablespaceList(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req schema.TablespaceListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the tablespaces
	response, err := manager.ListTablespaces(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func tablespaceCreate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req struct {
		Location string `json:"location"`
		schema.TablespaceMeta
	}
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the tablespace
	response, err := manager.CreateTablespace(r.Context(), req.TablespaceMeta, req.Location)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func tablespaceGet(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	// Get the tablespace
	response, err := manager.GetTablespace(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func tablespaceDelete(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	// Delete the tablespace
	_, err := manager.DeleteTablespace(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

func tablespaceUpdate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	// Parse request
	var req schema.TablespaceMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Update the tablespace
	response, err := manager.UpdateTablespace(r.Context(), name, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}
