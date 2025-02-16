package httpresponse

import (
	"encoding/json"
	"net/http"
	"strings"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// JSON will write a JSON response to the writer with a HTTP status
// code, optionally indenting the JSON by a number of spaces
func JSON(w http.ResponseWriter, code, indent int, v any) error {
	// Set the default content type, write the header
	w.Header().Set(types.ContentTypeHeader, types.ContentTypeJSON)
	w.WriteHeader(code)

	// Create an encoder for the response
	enc := json.NewEncoder(w)
	if indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", int(indent)))
	}

	// Write the JSON, and a trailing newline
	return enc.Encode(v)
}
