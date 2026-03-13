package httprouter

import (
	"net/http"
	"strings"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Cors returns an [HTTPMiddlewareFunc] that sets CORS response headers
// on every request. The origin parameter controls the Access-Control-Allow-Origin
// header:
//   - Empty string or "*": allow all origins (header value is set to "*")
//   - Any other value: used verbatim as the allowed origin
//
// headers lists the request headers that clients are permitted to send. If none
// are provided the header value defaults to "*". For preflight OPTIONS requests
// the middleware writes a 204 No Content response and does not call the next handler.
func Cors(origin string, headers ...string) HTTPMiddlewareFunc {
	allowOrigin := origin
	if allowOrigin == "" {
		allowOrigin = "*"
	}

	allowHeaders := "*"
	if len(headers) > 0 {
		var valid []string
		for _, h := range headers {
			if h = strings.TrimSpace(h); types.IsValidHeaderKey(h) {
				valid = append(valid, h)
			}
		}
		if len(valid) > 0 {
			allowHeaders = strings.Join(valid, ", ")
		}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			if r.Method == http.MethodOptions {
				// Echo back the requested method when present; fall back to "*".
				allowMethods := r.Header.Get("Access-Control-Request-Method")
				if allowMethods == "" {
					allowMethods = "*"
				}
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}
}
