package httprouter

import (
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	// Modules
	"github.com/djthorpe/go-marshaler"
	. "github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ErrorResponse struct {
	Code   int    `json:"code"`
	Reason string `json:"reason,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeJSON           = "application/json"
	ContentTypeText           = "text/plain"
	ContentTypeMultipartForm  = "multipart/form-data"
	ContentTypeUrlEncodedForm = "application/x-www-form-urlencoded"
)

const (
	// maxMemoryDefault to allocate for reading multipart forms
	maxMemoryDefault = 64 * 1024 // 64K
)

var (
	// type for url.Value parameters
	typeQueryValue = reflect.TypeOf([]string{})
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RequestParams returns the parameters encoded within the reqular expression
func RequestParams(req *http.Request) []string {
	return provider.ContextParams(req.Context())
}

// RequestQuery returns the query string of a request as a struct
func RequestQuery(req *http.Request, v interface{}) error {
	return decodeValues(req.URL.Query(), v)
}

// RequestBody returns the body of a request as a struct
func RequestBody(req *http.Request, v interface{}) error {
	if contentType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type")); err != nil {
		return err
	} else if contentType == ContentTypeJSON {
		return RequestBodyJSON(req, v)
	} else if contentType == ContentTypeMultipartForm {
		return RequestBodyForm(req, params, v)
	} else if contentType == ContentTypeUrlEncodedForm {
		return RequestBodyPost(req, v)
	} else {
		return ErrBadParameter.With(contentType)
	}
}

// RequestBodyJSON returns the body of a request as a struct
// where the request is of type application/json
func RequestBodyJSON(req *http.Request, v interface{}) error {
	reader := req.Body
	defer reader.Close()
	if err := json.NewDecoder(reader).Decode(v); err != nil {
		return err
	} else {
		return nil
	}
}

// RequestBodyForm returns the body of a request as a struct
// where the request is of type multipart/form-data
func RequestBodyForm(req *http.Request, params map[string]string, v interface{}) error {
	boundary, exists := params["boundary"]
	if !exists {
		return ErrBadParameter
	}
	defer req.Body.Close()

	if form, err := multipart.NewReader(req.Body, boundary).ReadForm(maxMemoryDefault); err != nil {
		return err
	} else {
		return decodeValues(form.Value, v)
	}
}

// RequestBodyPost returns the body of a request as a struct
// where the request is of type application/x-www-form-urlencoded
func RequestBodyPost(req *http.Request, v interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	} else {
		return decodeValues(req.PostForm, v)
	}
}

// ServeJSON is a utility function to serve an arbitary object as JSON
func ServeJSON(w http.ResponseWriter, v interface{}, code, indent uint) error {
	w.Header().Add("Content-Type", ContentTypeJSON)
	w.WriteHeader(int(code))
	enc := json.NewEncoder(w)
	if indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", int(indent)))
	}
	return enc.Encode(v)
}

// ServeText is a utility function to serve plaintext
func ServeText(w http.ResponseWriter, v string, code int) {
	w.Header().Add("Content-Type", ContentTypeText)
	w.WriteHeader(code)
	w.Write([]byte(v + "\n"))
}

// ServeError is a utility function to serve a JSON error notice
func ServeError(w http.ResponseWriter, code int, reason ...string) error {
	err := ErrorResponse{code, strings.Join(reason, " ")}
	if len(reason) == 0 {
		err.Reason = http.StatusText(code)
	}
	return ServeJSON(w, err, uint(code), 0)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func decodeValues(values url.Values, v interface{}) error {
	decoder := marshaler.NewDecoder("json", queryDecoder)
	q := make(map[string]interface{}, len(values))
	for k, v := range values {
		q[k] = v
	}
	return decoder.Decode(q, v)
}

func queryDecoder(src reflect.Value, dest reflect.Type) (reflect.Value, error) {
	// Source should be []string
	if src.Kind() != reflect.Slice && src.Type() != typeQueryValue {
		return reflect.ValueOf(nil), nil
	}
	// If there is no length to the query, then return zero-value
	if src.Len() == 0 {
		return reflect.Zero(dest), nil
	}
	// Decode first parameter
	switch dest.Kind() {
	case reflect.Uint:
		if value, err := strconv.ParseUint(src.Index(0).String(), 0, 64); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(uint(value)), nil
		}
	case reflect.Uint64:
		if value, err := strconv.ParseUint(src.Index(0).String(), 0, 64); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(value), nil
		}
	case reflect.Uint32:
		if value, err := strconv.ParseUint(src.Index(0).String(), 0, 32); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(uint32(value)), nil
		}
	case reflect.Int:
		if value, err := strconv.ParseInt(src.Index(0).String(), 0, 64); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(int(value)), nil
		}
	case reflect.Int64:
		if value, err := strconv.ParseInt(src.Index(0).String(), 0, 64); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(value), nil
		}
	case reflect.Int32:
		if value, err := strconv.ParseInt(src.Index(0).String(), 0, 32); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(int32(value)), nil
		}
	case reflect.String:
		return reflect.ValueOf(src.Index(0).String()), nil
	default:
		return reflect.ValueOf(nil), ErrInternalAppError.With("Unsupported type", dest.Kind())
	}
}
