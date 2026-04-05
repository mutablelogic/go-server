package httpresponse

import (
	"errors"
	"fmt"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Err turns a HTTP status code into an error
type Err int

// ErrResponse is the JSON error body returned by all error responses.
// It is also used to generate the OpenAPI response schema.
type ErrResponse struct {
	Code   int    `json:"code"             jsonschema:"HTTP status code"`
	Reason string `json:"reason,omitempty" jsonschema:"Human-readable reason phrase"`
	Detail any    `json:"detail,omitempty" jsonschema:"Optional additional detail"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ErrBadRequest         = Err(http.StatusBadRequest)
	ErrNotFound           = Err(http.StatusNotFound)
	ErrConflict           = Err(http.StatusConflict)
	ErrNotImplemented     = Err(http.StatusNotImplemented)
	ErrInternalError      = Err(http.StatusInternalServerError)
	ErrNotAuthorized      = Err(http.StatusUnauthorized)
	ErrForbidden          = Err(http.StatusForbidden)
	ErrServiceUnavailable = Err(http.StatusServiceUnavailable)
	ErrGatewayError       = Err(http.StatusBadGateway)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Error writes an error from a HTTP status code, with additional detail
func Error(w http.ResponseWriter, err error, detail ...any) error {
	// Create a JSON object for the error response
	e := ErrResponse{
		Code:   http.StatusInternalServerError,
		Reason: err.Error(),
	}

	// Modify the error code if it's a HTTP status code
	var code Err
	if errors.As(err, &code) {
		e.Code = int(code)
	}

	// Set the detail for the error
	if len(detail) == 1 {
		e.Detail = detail[0]
	} else if len(detail) > 1 {
		e.Detail = detail
	}

	// Write the error response
	return JSON(w, e.Code, 2, e)
}

///////////////////////////////////////////////////////////////////////////////
// ERROR

func (err ErrResponse) Error() string {
	return types.Stringify(err)
}

func (code Err) Error() string {
	return http.StatusText(int(code))
}

func (code Err) With(args ...interface{}) error {
	return fmt.Errorf("%w: %s", code, fmt.Sprint(args...))
}

func (code Err) Withf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", code, fmt.Sprintf(format, args...))
}
