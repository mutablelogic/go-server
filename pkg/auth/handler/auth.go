package handler

import (
	"context"
	"errors"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/auth"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterAuth(ctx context.Context, router server.HTTPRouter, prefix string, manager *auth.Manager) {
	router.HandleFunc(ctx, prefix, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodPost)

		// Handle method
		switch r.Method {
		case http.MethodPost:
			_ = authorize(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Get a token for a user. Note the user may not be live, but any token to
// make this request is valid.
func authorize(w http.ResponseWriter, r *http.Request, manager *auth.Manager) error {
	var value string

	// Get the token value from the post body
	if err := httprequest.Read(r, &value); err != nil {
		return httpresponse.Error(w, err)
	}

	// Return error if the token is empty
	if value == "" {
		return httpresponse.Error(w, httpresponse.Err(http.StatusUnauthorized))
	}

	// Get the user for the token value
	user, err := manager.GetUserForToken(r.Context(), value)
	if errors.Is(err, httpresponse.ErrNotFound) {
		return httpresponse.Error(w, httpresponse.Err(http.StatusUnauthorized))
	} else if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), user)
}
