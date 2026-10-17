package httprequest

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"

	// Packages
	jsonv1 "encoding/json"
	jsontext "encoding/json/jsontext"
	json "encoding/json/v2"
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
	case types.ContentTypeForm:
		return readFormURLEncoded(r, v)
	}

	// Cannot handle this content type
	return errBadRequest.Withf("unexpected content type %q", contentType)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// readJson decodes the JSON body into v, as encoding/json does, except that a
// time.Duration in v can be a duration string, such as "5s", as well as
// integer nanoseconds
func readJson(r *http.Request, v any) error {
	opts := json.JoinOptions(jsonv1.DefaultOptionsV1(), json.WithUnmarshalers(types.DurationUnmarshalers))
	if err := json.UnmarshalDecode(jsontext.NewDecoder(r.Body, opts), v, opts); errors.Is(err, io.EOF) {
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

	// Set non-file fields. A part without a filename is a value rather than a
	// file, so a value for a file field is set as a file without a name.
	values := make(url.Values, len(r.MultipartForm.Value))
	unnamed := make(map[string][]string)
	for key, vs := range r.MultipartForm.Value {
		if field, err := writableFieldForName(v, key); err == nil && (field.Type() == typeFile || field.Type() == typeFileSlice) {
			unnamed[key] = vs
		} else {
			values[key] = vs
		}
	}
	if err := Query(values, v); err != nil {
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

	// Set file fields from values, after any files with a name
	for key, vs := range unnamed {
		value, err := writableFieldForName(v, key)
		if err != nil {
			return err
		}
		switch value.Type() {
		case typeFile:
			if value.Interface().(types.File).Body == nil {
				value.Set(reflect.ValueOf(types.File{Body: newValueBody(vs[0])}))
			}
		case typeFileSlice:
			files := value.Interface().([]types.File)
			for _, v := range vs {
				files = append(files, types.File{Body: newValueBody(v)})
			}
			value.Set(reflect.ValueOf(files))
		}
	}

	// Return success
	return nil
}

// valueBody is the body of a file which was a value in a form, as its part
// had no filename, which can be read and seeked like a file's body
type valueBody struct {
	*strings.Reader
}

func newValueBody(value string) io.ReadCloser {
	return valueBody{strings.NewReader(value)}
}

func (valueBody) Close() error {
	return nil
}

func readFormURLEncoded(r *http.Request, v any) error {
	if err := r.ParseForm(); err != nil {
		return errBadRequest.With(err.Error())
	}
	if r.PostForm == nil {
		return httpresponse.ErrBadRequest.With("Missing form data")
	}
	if err := Query(r.PostForm, v); err != nil {
		return err
	}
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
