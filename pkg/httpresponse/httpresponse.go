package httpresponse

import (
	"encoding/json"
	"net/http"
	"strings"

	// Package imports
	err "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ErrorResponse is a generic error response which is served as JSON using the
// ServeError method
type ErrorResponse struct {
	Code   int    `json:"code"`
	Reason string `json:"reason,omitempty"`
	Detail any    `json:"detail,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeKey        = "Content-Type"
	ContentLengthKey      = "Content-Length"
	ContentTypeJSON       = "application/json"
	ContentTypeText       = "text/plain"
	ContentTypeTextStream = "text/event-stream"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// JSON is a utility function to serve an arbitary object as JSON. Setting
// the indent value to a non-zero value will indent the JSON. Additional
// header tuples can be provided as a series of key-value pairs
func JSON(w http.ResponseWriter, v interface{}, code int, indent uint, tuples ...string) error {
	if w == nil {
		return err.ErrBadParameter.With("nil response writer")
	}
	if len(tuples)%2 != 0 {
		return err.ErrBadParameter.With("odd number of tuples")
	}

	// Set the default content type
	w.Header().Set(ContentTypeKey, ContentTypeJSON)

	// Set additional headers
	for i := 0; i < len(tuples); i += 2 {
		w.Header().Set(tuples[i], tuples[i+1])
	}

	// Write the response
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	if indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", int(indent)))
	}
	return enc.Encode(v)
}

// Text is a utility function to serve plaintext. Additional header tuples can
// be provided as a series of key-value pairs
func Text(w http.ResponseWriter, v string, code int, tuples ...string) error {
	if w == nil {
		return err.ErrBadParameter.With("nil response writer")
	}
	if len(tuples)%2 != 0 {
		return err.ErrBadParameter.With("odd number of tuples")
	}

	// Set the default content type
	w.Header().Set(ContentTypeKey, ContentTypeText)

	// Set additional headers
	for i := 0; i < len(tuples); i += 2 {
		w.Header().Set(tuples[i], tuples[i+1])
	}

	// Write the response
	w.WriteHeader(code)
	if _, err := w.Write([]byte(v)); err != nil {
		return err
	}

	// If there is no trailing newline then write one
	if len(v) == 0 || v[len(v)-1] != '\n' {
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

// Empty is a utility function to serve an empty response. Additional header tuples can
// be provided as a series of key-value pairs
func Empty(w http.ResponseWriter, code int, tuples ...string) error {
	if w == nil {
		return err.ErrBadParameter.With("nil response writer")
	}
	if len(tuples)%2 != 0 {
		return err.ErrBadParameter.With("odd number of tuples")
	}

	// Set zero content length
	w.Header().Set(ContentLengthKey, "0")

	// Set additional headers
	for i := 0; i < len(tuples); i += 2 {
		w.Header().Set(tuples[i], tuples[i+1])
	}

	w.WriteHeader(code)

	// Return success
	return nil
}

// Error is a utility function to serve a JSON error notice
func Error(w http.ResponseWriter, code int, reason ...string) error {
	return ErrorWith(w, code, nil, reason...)
}

// ErrorWith is a utility function to serve a JSON error notice
// with additional details
func ErrorWith(w http.ResponseWriter, code int, v any, reason ...string) error {
	if w == nil {
		return err.ErrBadParameter.With("nil response writer")
	}
	if code == 0 {
		code = http.StatusInternalServerError
	}
	err := ErrorResponse{Code: code, Reason: strings.Join(reason, " "), Detail: v}
	if len(reason) == 0 {
		err.Reason = http.StatusText(int(code))
	}
	indent := uint(0)
	if v != nil {
		indent = 2
	}
	return JSON(w, err, code, indent)
}

// Cors is a utility function to set the CORS headers
// on a pre-flight request. Setting origin to an empty
// string or not including methods will allow any
// origin and any method
func Cors(w http.ResponseWriter, origin string, methods ...string) error {
	methods_ := "*"
	if len(methods) > 0 {
		methods_ = strings.ToUpper(strings.Join(methods, ","))
	}
	if origin == "" {
		origin = "*"
	}
	return Empty(w, http.StatusOK,
		"Access-Control-Allow-Origin", origin,
		"Access-Control-Allow-Methods", methods_,
		"Access-Control-Allow-Headers", "*",
	)
}
