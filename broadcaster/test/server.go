// Package test has a HTTP test server of the broadcaster's API, for the
// tests of the broadcaster, and of servers which use it
package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Server is a HTTP test server of an API
type Server struct {
	*httptest.Server
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Prefix is the path prefix of the API's routes
const Prefix = "/api"

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewServerFunc starts a HTTP test server of the API which the function
// registers, which is closed when the test ends
func NewServerFunc(t *testing.T, register func(*httprouter.Router) error) *Server {
	t.Helper()
	router, err := httprouter.NewRouter(context.Background(), http.NewServeMux(), Prefix, "*", "broadcaster", "test")
	if err != nil {
		t.Fatal(err)
	} else if err := register(router); err != nil {
		t.Fatal(err)
	}
	self := &Server{Server: httptest.NewServer(router)}
	t.Cleanup(self.Close)
	return self
}
