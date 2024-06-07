package certmanager

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"regexp"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqCreateCA struct {
	CommonName string `json:"name"`
	// Days       int    `json:"days"`
}

type reqCreateCert struct {
	reqCreateCA
	CA string `json:"ca"`
	// Hosts []string `json:"hosts"`
}

type respCert struct {
	Cert
	Certificate string `json:"certificate,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
	Error       string `json:"error,omitempty"`
}

// Check interfaces are satisfied
var _ server.ServiceEndpoints = (*certmanager)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reRoot   = regexp.MustCompile(`^/?$`)
	reCA     = regexp.MustCompile(`^/ca/?$`)
	reSerial = regexp.MustCompile(`^/([0-9]+)/?$`)
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

	// Path: /
	// Methods: POST
	// Scopes: write
	// Description: Create a new certificate
	r.AddHandlerFuncRe(ctx, reRoot, service.reqCreateCert, http.MethodPost).(router.Route).
		SetScope(service.ScopeWrite()...)

	// Path: /ca
	// Methods: POST
	// Scopes: write
	// Description: Create a new certificate authority
	r.AddHandlerFuncRe(ctx, reCA, service.reqCreateCA, http.MethodPost).(router.Route).
		SetScope(service.ScopeWrite()...)

	// Path: /<serial>
	// Methods: GET
	// Scopes: read
	// Description: Read a certificate by serial number
	r.AddHandlerFuncRe(ctx, reSerial, service.reqGetCert, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get all certificates
func (service *certmanager) reqListCerts(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, service.List(), http.StatusOK, jsonIndent)
}

// Get a certificate or CA
func (service *certmanager) reqGetCert(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())

	// Get the certificate
	cert, err := service.Read(urlParameters[0])
	if errors.Is(err, ErrNotFound) {
		httpresponse.Error(w, http.StatusNotFound, err.Error())
		return
	} else if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the certificate
	respCert := respCert{
		Cert: cert,
	}

	// Add any errors
	if !cert.IsCA() {
		if err := cert.IsValid(); err != nil {
			respCert.Error = err.Error()
		}
	}

	// Add public key
	var publicKey bytes.Buffer
	if err := cert.WriteCertificate(&publicKey); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
	} else {
		respCert.Certificate = publicKey.String()
	}

	// TODO: Add private key if scope allows

	httpresponse.JSON(w, respCert, http.StatusOK, jsonIndent)
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

// Create a new certificate
func (service *certmanager) reqCreateCert(w http.ResponseWriter, r *http.Request) {
	var req reqCreateCert

	// Get the request
	if err := httprequest.Read(r, &req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the CA
	var ca Cert
	if req.CA != "" {
		var err error
		ca, err = service.Read(req.CA)
		if errors.Is(err, ErrNotFound) {
			httpresponse.Error(w, http.StatusNotFound, err.Error())
			return
		} else if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Create the Cert
	if cert, err := service.CreateSignedCert(req.CommonName, ca); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
	} else {
		httpresponse.JSON(w, cert, http.StatusOK, jsonIndent)
	}
}
