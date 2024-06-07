package certmanager

import (
	"context"
	"net/http"
	"regexp"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	"github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqCreateCA struct {
	CommonName string `json:"name"`
}

// Check interfaces are satisfied
var _ server.ServiceEndpoints = (*certmanager)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reRoot = regexp.MustCompile(`^/?$`)
	reCA   = regexp.MustCompile(`^/ca/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *certmanager) AddEndpoints(ctx context.Context, r server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Return all existing certificates
	r.AddHandlerFuncRe(ctx, reRoot, service.reqListCerts, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /ca
	// Methods: POST
	// Scopes: write
	// Description: Create a new certificate authority
	r.AddHandlerFuncRe(ctx, reCA, service.reqCreateCA, http.MethodPost).(router.Route).
		SetScope(service.ScopeWrite()...)

}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get all certificates
func (service *certmanager) reqListCerts(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, service.List(), http.StatusOK, jsonIndent)
}

// Create a new certificate authority
func (service *certmanager) reqCreateCA(w http.ResponseWriter, r *http.Request) {
	var req reqCreateCA

	// Get the request
	if err := httprequest.Read(r, &req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create the CA
	if ca, err := service.CreateCA(req.CommonName); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
	} else {
		httpresponse.JSON(w, ca, http.StatusOK, jsonIndent)
	}
}
