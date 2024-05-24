package httpresponse

import (
	"encoding/json"
	"net/http"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ErrorResponse is a generic error response which is served as JSON using the
// ServeError method
type ErrorResponse struct {
	Code   uint   `json:"code"`
	Reason string `json:"reason,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeKey   = "Content-Type"
	ContentLengthKey = "Content-Length"
	ContentTypeJSON  = "application/json"
	ContentTypeText  = "text/plain"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// JSON is a utility function to serve an arbitary object as JSON
func JSON(w http.ResponseWriter, v interface{}, code, indent uint) error {
	if w == nil {
		return nil
	}
	w.Header().Add(ContentTypeKey, ContentTypeJSON)
	w.WriteHeader(int(code))
	enc := json.NewEncoder(w)
	if indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", int(indent)))
	}
	return enc.Encode(v)
}

// Text is a utility function to serve plaintext
func Text(w http.ResponseWriter, v string, code uint) {
	if w == nil {
		return
	}
	w.Header().Add(ContentTypeKey, ContentTypeText)
	w.WriteHeader(int(code))
	w.Write([]byte(v + "\n"))
}

// Empty is a utility function to serve an empty response
func Empty(w http.ResponseWriter, code uint) {
	if w == nil {
		return
	}
	w.Header().Add(ContentLengthKey, "0")
	w.WriteHeader(int(code))
}

// Error is a utility function to serve a JSON error notice
func Error(w http.ResponseWriter, code uint, reason ...string) error {
	if w == nil {
		return nil
	}
	err := ErrorResponse{code, strings.Join(reason, " ")}
	if len(reason) == 0 {
		err.Reason = http.StatusText(int(code))
	}
	return JSON(w, err, code, 0)
}
