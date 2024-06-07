package certmanager

import (
	"context"
	"net/http"
	"regexp"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.ServiceEndpoints = (*certmanager)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reRoot = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *certmanager) AddEndpoints(ctx context.Context, r server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Return all existing certificates
	r.AddHandlerFuncRe(ctx, reRoot, service.ListCerts, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get all certificates
func (service *certmanager) ListCerts(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, service.List(), http.StatusOK, jsonIndent)
}
