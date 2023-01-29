package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	// Module imports
	"github.com/mutablelogic/go-server/pkg/httpserver/util"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Register middleware for tokenauth
func (tokenauth *tokenauth) TokenAuthMiddleware(child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer, err := bearer(r.Header.Get(authHeader), authScheme)
		if err != nil {
			util.ServeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		fmt.Println("bearer", bearer)
		// Continue with the child handler
		child(w, r)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// bearer returns the bearer token from the authorization header, which should
// be base64 encoded
func bearer(value, scheme string) (string, error) {
	if value == "" {
		return "", ErrBadParameter.With("Missing authorization header")
	}
	bearer := strings.TrimPrefix(value, scheme)
	if bearer == "" || bearer == value {
		return "", ErrBadParameter.With("Missing authorization value")
	} else if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(bearer)); err != nil {
		return "", err
	} else {
		return string(decoded), nil
	}
}
