package handler

import (
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ldap "github.com/mutablelogic/go-server/pkg/ldap"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func userList(w http.ResponseWriter, r *http.Request, manager *ldap.Manager) error {
	// Parse request
	var req schema.ObjectListRequest
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

func userCreate(w http.ResponseWriter, r *http.Request, manager *ldap.Manager) error {
	// Parse request
	var req schema.Object
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the object
	response, err := manager.CreateUser(r.Context(), req.DN, req.Values)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func userGet(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, user string) error {
	// Get the user
	response, err := manager.GetUser(r.Context(), user)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func userDelete(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, user string) error {
	// Delete the y=user
	_, err := manager.DeleteUser(r.Context(), user)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}
