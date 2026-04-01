package httprouter

import (
	"net/http"

	// Packages
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type SecurityScheme interface {
	// Wrap returns a HTTP handler function that wraps the security scheme around
	// the provided handler, enforcing the required scopes.
	Wrap(handler http.HandlerFunc, scopes []string) http.HandlerFunc

	// Spec returns the openapi.PathItem for a path with optional path parameters
	Spec() openapi.SecurityScheme
}
