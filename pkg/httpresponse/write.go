package httpresponse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write a custom response to the writer with a HTTP status code. If the custom
// function is set and returns the number of bytes written (greater than zero), it will be used
// to write the response body.
func Write(w http.ResponseWriter, code int, contentType string, fn func(w io.Writer) (int, error)) error {
	if w.Header().Get(types.ContentTypeHeader) == "" {
		w.Header().Set(types.ContentTypeHeader, contentType)
	}

	// Modify the status code if it is not already set
	if code == 0 {
		code = http.StatusOK
	}

	// If fn is not nil, call it to write the response body
	var buf bytes.Buffer
	if fn != nil {
		n, err := fn(&buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		if n > 0 {
			w.Header().Set(types.ContentLengthHeader, fmt.Sprint(n))
		}
	}

	// Write the status code
	w.WriteHeader(code)

	// Write data
	if buf.Len() > 0 {
		if _, err := w.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	// Return success
	return nil
}
