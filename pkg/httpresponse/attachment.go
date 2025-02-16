package httpresponse

import (
	"io"
	"mime"
	"net/http"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Attachment will write out an attachment from a io.Reader
func Attachment(w http.ResponseWriter, r io.Reader, code int, params map[string]string) error {
	// Set the default content type to binary
	if w.Header().Get(types.ContentTypeHeader) == "" {
		w.Header().Set(types.ContentTypeHeader, types.ContentTypeBinary)
	}

	// Set the content disposition
	w.Header().Set(types.ContentDispositonHeader, mime.FormatMediaType("attachment", params))

	// Set the status
	w.WriteHeader(code)

	// Copy the reader to the writer
	_, err := io.Copy(w, r)
	return err
}
