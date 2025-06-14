package httprequest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"

	// Packages
	gomultipart "github.com/mutablelogic/go-client/pkg/multipart"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	FormDataMaxMemory = 256 << 20 // 256 MB
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
	case types.ContentTypeFormData:
		return readFormData(r, v)
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

var (
	typeFile = reflect.TypeOf(gomultipart.File{})
)

func readFormData(r *http.Request, v any) error {
	if err := r.ParseMultipartForm(FormDataMaxMemory); err != nil {
		return err
	}
	if r.MultipartForm == nil {
		return httpresponse.ErrBadRequest.With("Missing form data")
	}

	// Set non-file fields
	if err := Query(r.MultipartForm.Value, v); err != nil {
		return err
	}
	// Set file fields - we only support one file per field
	for key, values := range r.MultipartForm.File {
		if len(values) == 0 {
			continue
		}
		// Get the first file for the field
		value, err := writableFieldForName(v, key)
		if err != nil {
			return err
		}
		switch value.Type() {
		case typeFile:
			body, err := values[0].Open()
			if err != nil {
				return errBadRequest.Withf("cannot open file %q: %v", values[0].Filename, err)
			}
			value.Set(reflect.ValueOf(gomultipart.File{
				Path: values[0].Filename,
				Body: body,
			}))
		default:
			return httpresponse.ErrBadRequest.Withf("cannot set field %q of type %s", key, value.Type())
		}
	}

	// Return success
	return nil
}
