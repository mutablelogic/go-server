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

func RegisterSchema(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgmanager.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "schema"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodGet:
			_ = schemaList(w, r, manager)
		case http.MethodPost:
			_ = schemaCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "schema/{name...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		// Parse request argument
		name := r.PathValue("name")

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = schemaGet(w, r, manager, name)
		case http.MethodDelete:
			_ = schemaDelete(w, r, manager, name)
		case http.MethodPatch:
			_ = schemaUpdate(w, r, manager, name)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func schemaList(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req schema.SchemaListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the schemas
	response, err := manager.ListSchemas(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func schemaCreate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req schema.SchemaMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the schema
	response, err := manager.CreateSchema(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func schemaGet(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	schema, err := manager.GetSchema(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), schema)
}

func schemaDelete(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	// Parse the query
	var req struct {
		Force bool `json:"force,omitempty" help:"Force delete"`
	}
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Delete the schema
	_, err := manager.DeleteSchema(r.Context(), name, req.Force)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

func schemaUpdate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, name string) error {
	// Parse request
	var req schema.SchemaMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Perform update
	schema, err := manager.UpdateSchema(r.Context(), name, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), schema)
}
