package httprequest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read the request body into a structure
func Read(r *http.Request, v interface{}) error {
	// Determine the content type
	contentType, err := types.RequestContentType(r)
	if err != nil {
		return errBadRequest.With(err.Error())
	}

	// Read the body according to the content type
	switch contentType {
	case types.ContentTypeJSON:
		return readJson(r, v)
	}

	// Cannot handle this content type
	return errBadRequest.Withf("unexpected content type %q", contentType)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readJson(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); errors.Is(err, io.EOF) {
		return errBadRequest.With("Missing request body")
	} else if err != nil {
		return errBadRequest.With(err.Error())
	}
	return nil
}
