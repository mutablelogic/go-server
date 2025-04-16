package auth

import (
	"net/http"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

/////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ApiKeyHeader = "X-API-Key"
	ApiKeyQuery  = "api-key"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

// The AuthFunc should return a httpresponse error, or nil if the user is
// authorized to carry out a request
type AuthFunc func(*schema.User) error

type Auth interface {
	// Wrap a http.Handler with an authentication function, calling the function
	// to do the authorization check
	WrapFunc(http.HandlerFunc, AuthFunc) http.HandlerFunc
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns the API token from the request, checking for the X-API-Key header first
// and then the api-key query parameter (not recommended)
func TokenForRequest(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get(ApiKeyHeader)); token != "" {
		return token
	}
	if token := strings.TrimSpace(r.URL.Query().Get(ApiKeyQuery)); token != "" {
		return token
	}
	return ""
}
