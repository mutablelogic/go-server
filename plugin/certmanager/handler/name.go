package handler

import (
	"context"
	"net/http"

	// Packages

	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/cert"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterName(ctx context.Context, router server.HTTPRouter, prefix string, certmanager *cert.CertManager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "name"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet)

		switch r.Method {
		case http.MethodGet:
			_ = nameList(w, r, certmanager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func nameList(w http.ResponseWriter, r *http.Request, certmanager *cert.CertManager) error {
	response, err := certmanager.ListNames(r.Context())
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
