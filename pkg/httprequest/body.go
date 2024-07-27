package httprequest

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"slices"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeJson           = "application/json"
	ContentTypeTextXml        = "text/xml"
	ContentTypeApplicationXml = "application/xml"
	ContentTypeText           = "text/"
	ContentTypeBinary         = "application/octet-stream"
	ContentTypeFormData       = "multipart/form-data"
	ContentTypeUrlEncoded     = "application/x-www-form-urlencoded"
)

const (
	maxMemory = 10 << 20 // 10 MB in-memory cache for multipart form
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Body reads the body of an HTTP request and decodes it into v.
// You can include the mimetypes that are acceptable, otherwise it will read
// the body based on the content type. If does not close the body of the
// request, which should be done by the caller.
// v can be a io.Writer, in which case the body is copied into it.
// v can be a struct, in which case the body is decoded into it.
func Body(v any, r *http.Request, accept ...string) error {
	// Parse the content type
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	} else {
		contentType = strings.ToLower(contentType)
	}

	// Check whether we'll accept it
	if len(accept) > 0 && !slices.Contains(accept, contentType) {
		return ErrBadParameter.Withf("unexpected content type %q", contentType)
	}

	// If v is an io.Writer, then copy the body into it
	if v, ok := v.(io.Writer); ok {
		if _, err := io.Copy(v, r.Body); err != nil {
			return err
		}
		return nil
	}

	// Read the body - we can read JSON, form data or url encoded data
	// TODO: We should also be able to read XML
	switch {
	case contentType == ContentTypeJson:
		return readJson(v, r)
	case contentType == ContentTypeFormData:
		return readFormData(v, r)
	case contentType == ContentTypeUrlEncoded:
		return readForm(v, r)
	case strings.HasPrefix(contentType, ContentTypeText):
		return readText(v, r)
	}

	// Report error
	return ErrBadParameter.Withf("unsupported content type %q", contentType)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readJson(v any, r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func readForm(v any, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	} else {
		return Query(v, r.Form)
	}
}

func readFormData(v any, r *http.Request) error {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return err
	} else if err := readFiles(v, r.MultipartForm.File); err != nil {
		return err
	} else {
		return Query(v, r.MultipartForm.Value)
	}
}

func readFiles(v any, files map[string][]*multipart.FileHeader) error {
	return mapFields(v, func(key string, value reflect.Value) error {
		src, exists := files[key]
		if !exists || len(src) == 0 {
			setZeroValue(value)
			return nil
		}
		// Set a *multipart.FileHeader or []*multipart.FileHeader with the source
		return setFile(value, src)
	})
}

func readText(v any, r *http.Request) error {
	switch v := v.(type) {
	case *string:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		*v = string(data)
	default:
		return ErrBadParameter.Withf("unsupported type: %T", v)
	}

	// Return success
	return nil
}
