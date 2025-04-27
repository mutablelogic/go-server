package handler

import (
	"context"
	"net/http"
	"strconv"

	// Packages
	server "github.com/mutablelogic/go-server"
	cert "github.com/mutablelogic/go-server/pkg/cert"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterName(ctx context.Context, router server.HTTPRouter, prefix string, certmanager *cert.CertManager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "name"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodGet:
			_ = nameList(w, r, certmanager)
		case http.MethodPost:
			_ = nameCreate(w, r, certmanager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "name/{id}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodDelete, http.MethodGet, http.MethodPatch)

		// Parse request
		id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest, r.PathValue("id"))
			return
		}

		// Handle request
		switch r.Method {
		case http.MethodGet:
			_ = nameGet(w, r, certmanager, id)
		case http.MethodPatch:
			_ = nameUpdate(w, r, certmanager, id)
		case http.MethodDelete:
			_ = nameDelete(w, r, certmanager, id)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func nameList(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager) error {
	// Parse request
	var req schema.NameListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	response, err := certmanager.ListNames(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func nameDelete(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager, req uint64) error {
	_, err := certmanager.DeleteName(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

func nameCreate(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager) error {
	// Parse request
	var req schema.NameMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the name
	response, err := certmanager.CreateName(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func nameGet(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager, req uint64) error {
	response, err := certmanager.GetName(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func nameUpdate(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager, req uint64) error {
	// Parse request
	var meta schema.NameMeta
	if err := httprequest.Read(r, &meta); err != nil {
		return httpresponse.Error(w, err)
	}

	// Update the name
	response, err := certmanager.UpdateName(r.Context(), req, meta)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
