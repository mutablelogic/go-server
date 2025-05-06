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

func objectList(w http.ResponseWriter, r *http.Request, manager *ldap.Manager) error {
	// Parse request
	var req schema.ObjectListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the objects
	response, err := manager.List(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func objectCreate(w http.ResponseWriter, r *http.Request, manager *ldap.Manager) error {
	// Parse request
	var req schema.Object
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the object
	response, err := manager.Create(r.Context(), req.DN, req.Values)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func objectGet(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, dn string) error {
	// Get the object
	response, err := manager.Get(r.Context(), dn)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func objectDelete(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, dn string) error {
	// Delete the object
	_, err := manager.Delete(r.Context(), dn)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.Empty(w, http.StatusOK)
}

func objectUpdate(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, dn string) error {
	// Parse request
	var req schema.Object
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Update the attributes
	// TODO: Need to support adding and removing attributes, and changing the DN
	response, err := manager.Update(r.Context(), dn, req.Values)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
