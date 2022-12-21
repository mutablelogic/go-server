package client

import (
	"net/http"
	"path/filepath"
	"strings"
)

// OptPath appends path elements onto a request
func OptPath(value ...string) RequestOpt {
	return func(r *http.Request) error {
		// Make a copy
		url := *r.URL
		// Clean up and append path
		url.Path = PathSeparator + filepath.Join(strings.Trim(url.Path, PathSeparator), strings.Join(value, PathSeparator))
		// Set new path
		r.URL = &url
		return nil
	}
}

// OptToken adds an authorization header. The header format is "Authorization: Bearer <token>"
func OptToken(value string) RequestOpt {
	return func(r *http.Request) error {
		if value != "" {
			r.Header.Set("Authorization", "Bearer "+value)
		}
		return nil
	}
}
