package handler

import (
	"context"
	"net/http"
	"strings"

	// Packages
	"github.com/djthorpe/go-pg"
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
			_ = schemaList(w, r, manager, nil)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "schema/{database}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		// Parse path argument
		database := strings.TrimSpace(r.PathValue("database"))
		if database == "" {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("database is required"))
		}

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = schemaList(w, r, manager, types.StringPtr(database))
		case http.MethodPost:
			_ = schemaCreate(w, r, manager, database)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "schema/{database}/{schema}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		// Parse path arguments
		database := strings.TrimSpace(r.PathValue("database"))
		if database == "" {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("database is required"))
		}
		namespace := strings.TrimSpace(r.PathValue("schema"))
		if namespace == "" {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("schema is required"))
		}

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = schemaGet(w, r, manager, database, namespace)
		case http.MethodDelete:
			_ = schemaDelete(w, r, manager, database, namespace)
		case http.MethodPatch:
			_ = schemaUpdate(w, r, manager, database, namespace)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// List all schemas in all databases
func schemaList(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, database *string) error {
	// Parse request
	var req pg.OffsetLimit
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the schemas
	response, err := manager.ListSchemas(r.Context(), schema.SchemaListRequest{
		Database:    database,
		OffsetLimit: req,
	})
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func schemaCreate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, database string) error {
	// Parse request
	var req schema.SchemaMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the schema
	response, err := manager.CreateSchema(r.Context(), database, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func schemaGet(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, database, namespace string) error {
	schema, err := manager.GetSchema(r.Context(), database, namespace)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), schema)
}

func schemaDelete(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, database, namespace string) error {
	// Parse the query
	var req struct {
		Force bool `json:"force,omitempty" help:"Force delete"`
	}
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Delete the schema
	_, err := manager.DeleteSchema(r.Context(), database, namespace, req.Force)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

func schemaUpdate(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, database, namespace string) error {
	// Parse request
	var req schema.SchemaMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Perform update
	schema, err := manager.UpdateSchema(r.Context(), database, namespace, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), schema)
}
