package httprouter

import (
	"fmt"
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

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterSecurityScheme registers a security scheme with the router. The scheme
// can then be referenced in the OpenAPI spec by name.
func (r *Router) RegisterSecurityScheme(name string, scheme SecurityScheme) error {
	// TODO
	fmt.Println("Registering securityscheme (TODO)")
	return nil
}
