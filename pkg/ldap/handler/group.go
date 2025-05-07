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

func groupList(w http.ResponseWriter, r *http.Request, manager *ldap.Manager) error {
	// Parse request
	var req schema.ObjectListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the groups
	response, err := manager.ListGroups(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func groupCreate(w http.ResponseWriter, r *http.Request, manager *ldap.Manager) error {
	// Parse request
	var req schema.Object
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the object
	response, err := manager.CreateGroup(r.Context(), req.DN, req.Values)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func groupGet(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, group string) error {
	// Get the group
	response, err := manager.GetGroup(r.Context(), group)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func groupDelete(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, group string) error {
	// Delete the y=group
	_, err := manager.DeleteGroup(r.Context(), group)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}
