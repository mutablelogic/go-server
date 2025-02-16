package httpresponse

import (
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Empty is a utility function to serve an empty response
func Empty(w http.ResponseWriter, code int) error {
	w.Header().Set(types.ContentLengthHeader, "0")
	w.WriteHeader(code)

	// Return success
	return nil
}
