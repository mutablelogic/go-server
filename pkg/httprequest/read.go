package httprequest

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"reflect"
	"strings"

	// Packages
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
	case *[]byte:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return httpresponse.ErrBadRequest.With(err.Error())
		}
		*v = data
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
	typeFile      = reflect.TypeOf(types.File{})
	typeFileSlice = reflect.TypeOf([]types.File{})
)

func readFormData(r *http.Request, v any) error {
	// ParseMultipartForm reads the entire request body eagerly: parts up to
	// FormDataMaxMemory bytes are held in memory, larger parts spill to OS temp
	// files. Cleanup (RemoveAll) is handled automatically: each opened file body
	// is wrapped in a refCountedBody, and RemoveAll is called once the last body
	// is closed by the caller.
	if err := r.ParseMultipartForm(FormDataMaxMemory); err != nil {
		return err
	}
	if r.MultipartForm == nil {
		return httpresponse.ErrBadRequest.With("Missing form data")
	}

	cleanup := &multipartCleanup{form: r.MultipartForm}
	defer cleanup.done() // calls RemoveAll immediately if no files were opened

	// Set non-file fields
	if err := Query(r.MultipartForm.Value, v); err != nil {
		return err
	}

	// Set file fields — supports both a single gomultipart.File and []gomultipart.File
	for key, values := range r.MultipartForm.File {
		if len(values) == 0 {
			continue
		}
		value, err := writableFieldForName(v, key)
		if err != nil {
			return err
		}
		switch value.Type() {
		case typeFile:
			// Backward-compatible single-file: use the first part only.
			body, err := cleanup.open(values[0])
			if err != nil {
				return errBadRequest.Withf("cannot open file %q: %v", values[0].Filename, err)
			}
			value.Set(reflect.ValueOf(types.File{
				Path:        fileHeaderPath(values[0]),
				Body:        body,
				ContentType: values[0].Header.Get("Content-Type"),
				Header:      values[0].Header,
			}))
		case typeFileSlice:
			// Multi-file: open every part and collect into a slice.
			files := make([]types.File, 0, len(values))
			for _, fh := range values {
				body, err := cleanup.open(fh)
				if err != nil {
					// Close any already-opened bodies; their Close() calls
					// decrement the ref count and will trigger RemoveAll when
					// the count reaches zero. Join all close errors.
					errs := []error{errBadRequest.Withf("cannot open file %q: %v", fh.Filename, err)}
					for _, f := range files {
						if cerr := f.Body.Close(); cerr != nil {
							errs = append(errs, cerr)
						}
					}
					return errors.Join(errs...)
				}
				files = append(files, types.File{
					Path:        fileHeaderPath(fh),
					Body:        body,
					ContentType: fh.Header.Get("Content-Type"),
					Header:      fh.Header,
				})
			}
			value.Set(reflect.ValueOf(files))
		default:
			return httpresponse.ErrBadRequest.Withf("cannot set field %q of type %s", key, value.Type())
		}
	}

	// Return success
	return nil
}

// fileHeaderPath returns the sanitised logical path for a multipart file part.
// It prefers the X-Path header (set by go-client when the path contains
// directory components, since the stdlib strips directories from the
// Content-Disposition filename per RFC 7578 §4.2) and falls back to
// the base filename exposed by the stdlib parser.
//
// The X-Path value is normalised with path.Clean and stripped of any leading
// slash. If the result is empty or escapes its root via .. segments, the
// function falls back to fh.Filename (already basename-only per the stdlib)
// to prevent path traversal attacks.
func fileHeaderPath(fh *multipart.FileHeader) string {
	p := fh.Header.Get(types.ContentPathHeader)
	if p == "" {
		return fh.Filename
	}
	// Resolve . and .. then strip any leading slash.
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	// If the cleaned path escapes the root or is degenerate, fall back to
	// the stdlib-provided basename which is always safe.
	if p == "." || p == "" || strings.HasPrefix(p, "..") {
		return fh.Filename
	}
	return p
}
