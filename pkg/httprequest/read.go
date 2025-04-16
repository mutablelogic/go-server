package httprequest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
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
	case types.ContentTypeTextPlain:
		return readString(r, v)
	}

	// Cannot handle this content type
	return errBadRequest.Withf("unexpected content type %q", contentType)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readJson(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); errors.Is(err, io.EOF) {
		return errBadRequest.With("Missing request body")
	} else if err != nil {
		return errBadRequest.With(err.Error())
	}
	return nil
}

func readString(r *http.Request, v any) error {
	switch v := v.(type) {
	case []byte:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return httpresponse.ErrBadRequest.With(err.Error())
		}
		copy(v, data)
		return nil
	case *string:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return httpresponse.ErrBadRequest.With(err.Error())
		}
		*v = string(data)
		return nil
	default:
		return httpresponse.ErrInternalError.Withf("cannot read %T as string", v)
	}
}
