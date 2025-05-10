package httpresponse

import (
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write a custom response to the writer with a HTTP status code,
// leave the actual writing of the response to the caller
func Write(w http.ResponseWriter, code int, contentType string) {
	if w.Header().Get(types.ContentTypeHeader) == "" {
		w.Header().Set(types.ContentTypeHeader, contentType)
	}

	// Modify the status code if it is not already set
	if code == 0 {
		code = http.StatusOK
	}

	// Write the status code
	w.WriteHeader(code)
}
