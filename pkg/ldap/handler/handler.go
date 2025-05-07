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
	registerAuth(ctx, router, schema.APIPrefix, manager)
	registerUser(ctx, router, schema.APIPrefix, manager)
	registerGroup(ctx, router, schema.APIPrefix, manager)
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
		case http.MethodPost:
			_ = objectCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "object/{dn...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = objectGet(w, r, manager, r.PathValue("dn"))
		case http.MethodDelete:
			_ = objectDelete(w, r, manager, r.PathValue("dn"))
		case http.MethodPatch:
			_ = objectUpdate(w, r, manager, r.PathValue("dn"))
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

func registerAuth(ctx context.Context, router server.HTTPRouter, prefix string, manager *ldap.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "auth/{dn...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodPost, http.MethodPut)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodPost:
			_ = authBind(w, r, manager, r.PathValue("dn"))
		case http.MethodPut:
			_ = authChangePassword(w, r, manager, r.PathValue("dn"))
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

func registerUser(ctx context.Context, router server.HTTPRouter, prefix string, manager *ldap.Manager) {
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

	router.HandleFunc(ctx, types.JoinPath(prefix, "user/{user...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete)

		// Get user parameter
		user := r.PathValue("user")
		if !types.IsIdentifier(user) {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusBadRequest), "invalid user")
			return
		}

		// Handle request
		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = userGet(w, r, manager, user)
		case http.MethodDelete:
			_ = userDelete(w, r, manager, user)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

func registerGroup(ctx context.Context, router server.HTTPRouter, prefix string, manager *ldap.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "group"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodGet:
			_ = groupList(w, r, manager)
		case http.MethodPost:
			_ = groupCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "group/{group...}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete)

		// Get group parameter
		group := r.PathValue("group")
		if !types.IsIdentifier(group) {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusBadRequest), "invalid group")
			return
		}

		// Handle request
		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = groupGet(w, r, manager, group)
		case http.MethodDelete:
			_ = groupDelete(w, r, manager, group)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}
