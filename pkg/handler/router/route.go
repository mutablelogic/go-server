package router

import (
	"encoding/json"
	"net/http"

	"github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Router interface {
	server.Router

	// Match a host, method and path to a handler. Returns the appropriate
	// http status code, which will be 200 on success, 404 or 405 and
	// path parameters extracted from the path.
	Match(host, method, path string) (*Route, int)
}

// Route is a matched route handler, with associated host, prefix and path
type Route struct {
	// Key
	Key string `json:"key,omitempty"`

	// Matched handler
	Handler http.HandlerFunc `json:"-"`

	// Matched host
	Host string `json:"host,omitempty"`

	// Matched prefix
	Prefix string `json:"prefix,omitempty"`

	// Matched path
	Path string `json:"path,omitempty"`

	// Computed	parameters from the path
	Parameters []string `json:"parameters,omitempty"`

	// Whether the result was from the cache
	Cached bool `json:"cached,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Route) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
