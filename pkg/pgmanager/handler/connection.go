package handler

import (
	"context"
	"net/http"
	"strconv"

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

func RegisterConnection(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgmanager.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "connection"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet)

		switch r.Method {
		case http.MethodGet:
			_ = connectionList(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "connection/{pid}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete)

		// Parse request argument
		pid, err := strconv.ParseUint(r.PathValue("pid"), 10, 64)
		if err != nil {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing or invalid pid"), r.PathValue("pid"))
			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = connectionGet(w, r, manager, pid)
		case http.MethodDelete:
			_ = connectionDelete(w, r, manager, pid)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func connectionList(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager) error {
	// Parse request
	var req schema.ConnectionListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the connections
	response, err := manager.ListConnections(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func connectionGet(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, pid uint64) error {
	// Get the connections
	response, err := manager.GetConnection(r.Context(), pid)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func connectionDelete(w http.ResponseWriter, r *http.Request, manager *pgmanager.Manager, pid uint64) error {
	// Delete the connections
	_, err := manager.DeleteConnection(r.Context(), pid)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}
