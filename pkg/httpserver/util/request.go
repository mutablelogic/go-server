package util

import (
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/context"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return prefix and parameters from a request
func ReqPrefixPathParams(req *http.Request) (string, string, []string) {
	return context.PrefixPathParams(req.Context())
}

// Return true if the admin flag is set
func ReqAdmin(req *http.Request) bool {
	return context.Admin(req.Context())
}
