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

func authBind(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, dn string) error {
	// Get the password from the post body
	var password string
	if err := httprequest.Read(r, &password); err != nil {
		return httpresponse.Error(w, err)
	}

	// Return error if the password is empty
	if password == "" {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing password"))
	}

	// Bind the password
	response, err := manager.Bind(r.Context(), dn, password)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func authChangePassword(w http.ResponseWriter, r *http.Request, manager *ldap.Manager, dn string) error {
	// Get the new password from the post body
	var password string
	if err := httprequest.Read(r, &password); err != nil {
		return httpresponse.Error(w, err)
	}

	// Change the password
	var response struct {
		*schema.Object
		Password string `json:"password,omitempty"`
	}

	// Set the new password
	response.Password = password

	// Change the password
	user, err := manager.ChangePassword(r.Context(), dn, "", &response.Password)
	if err != nil {
		return httpresponse.Error(w, err)
	} else {
		response.Object = user
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
