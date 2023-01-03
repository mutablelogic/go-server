package util

import (
	"mime"
	//"mime/multipart"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/context"
	"github.com/mutablelogic/go-server/pkg/httpserver/unmarshal"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return prefix and parameters from a request
func ReqPrefixPathParams(req *http.Request) (string, string, []string) {
	return context.PrefixPathParams(req.Context())
}

// Decode a request body into a struct, and return the fields
// which were decoded.
func ReqDecodeBody(req *http.Request, v interface{}) ([]string, error) {
	if contentType, params, err := mime.ParseMediaType(req.Header.Get(ContentTypeKey)); err != nil {
		return nil, err
	} else if contentType == ContentTypeJSON {
		return RequestBodyJSON(req, v)
	} else if contentType == ContentTypeMultipartForm {
		return RequestBodyForm(req, params, v)
	} else if contentType == ContentTypeUrlEncodedForm {
		return RequestBodyPost(req, v)
	} else {
		return nil, ErrBadParameter.With(contentType)
	}
}

// RequestBodyJSON returns the body of a request as a struct
// where the request is of type application/json
func RequestBodyJSON(req *http.Request, v interface{}) ([]string, error) {
	defer req.Body.Close()
	return unmarshal.WithJson(req.Body, v)
}

// RequestBodyForm returns the body of a request as a struct
// where the request is of type multipart/form-data
func RequestBodyForm(req *http.Request, params map[string]string, v interface{}) ([]string, error) {
	defer req.Body.Close()

	//boundary, exists := params["boundary"]
	//if !exists {
	//	return nil, ErrBadParameter
	//}

	//form, err := multipart.NewReader(req.Body, boundary).ReadForm(maxMemoryDefault)
	//if err != nil {
	//	return nil, err
	//}

	return nil, ErrNotImplemented
	// return unmarshal.WithForm(form, v)
}

// RequestBodyPost returns the body of a request as a struct
// where the request is of type application/x-www-form-urlencoded
func RequestBodyPost(req *http.Request, v interface{}) ([]string, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}
	return unmarshal.WithQuery(req.PostForm, v)
}
