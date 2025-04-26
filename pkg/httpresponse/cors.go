package httpresponse

import (
	"net/http"
	"strings"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

// Write the cors headers
func Cors(w http.ResponseWriter, r *http.Request, origin string, methods ...string) {
	if origin == "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	// Preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", corsMethods(methods...))
		w.Header().Set("Access-Control-Allow-Headers", corsHeaders(methods...))
	}
}

// Return a string of methods, which are in uppercase
func corsMethods(args ...string) string {
	var methods []string
	for _, method := range args {
		if method := strings.TrimSpace(method); types.IsUppercase(method) && method != "" {
			methods = append(methods, method)
		}
	}
	if len(methods) == 0 {
		return "*"
	}
	return strings.Join(methods, ", ")
}

// Return a string of headers, which are not in uppercase
func corsHeaders(args ...string) string {
	if len(args) == 0 {
		return "*"
	}
	var headers []string
	for _, header := range args {
		if header := strings.TrimSpace(header); !types.IsUppercase(header) && header != "" {
			headers = append(headers, header)
		}
	}
	if len(headers) == 0 {
		return "*"
	}
	return strings.Join(headers, ", ")
}
