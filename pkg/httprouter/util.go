package httprouter

import (
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	// Packages
	marshaler "github.com/djthorpe/go-marshaler"
	provider "github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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

func decodeValues(q url.Values, v interface{}) error {
	// Add decoders for duration and time values
	decoder := marshaler.NewDecoder("json", marshaler.ConvertQueryValues, marshaler.ConvertStringToNumber, marshaler.ConvertDuration, marshaler.ConvertTime)
	return decoder.DecodeQuery(q, v)
}
