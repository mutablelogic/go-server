package handler

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ldap "github.com/mutablelogic/go-server/pkg/ldap"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Register(ctx context.Context, router server.HTTPRouter, manager *ldap.Manager) {
	registerObject(ctx, router, schema.APIPrefix, manager)
}

func registerObject(ctx context.Context, router server.HTTPRouter, prefix string, manager *ldap.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "object"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = objectList(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}
