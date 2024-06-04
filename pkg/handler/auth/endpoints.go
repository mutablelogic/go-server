package auth

import (
	"context"
	"net/http"
	"regexp"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2

	// Token should be at least eight bytes (16 chars)
	reTokenString = `[a-zA-Z0-9]{16}[a-zA-Z0-9]*`
)

var (
	reRoot  = regexp.MustCompile(`^/?$`)
	reToken = regexp.MustCompile(`^/(` + reTokenString + `)/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - ENDPOINTS

// Add endpoints to the router
func (service *auth) AddEndpoints(ctx context.Context, router server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Get current set of tokens and groups
	router.AddHandlerFuncRe(ctx, reRoot, service.ListTokens, http.MethodGet)

	// Path: /
	// Methods: POST
	// Scopes: write // TODO: Add scopes
	// Description: Create a new token
	router.AddHandlerFuncRe(ctx, reRoot, service.CreateToken, http.MethodPost)

	// Path: /<token>
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Get a token
	router.AddHandlerFuncRe(ctx, reToken, service.GetToken, http.MethodGet)

	// Path: /<token>
	// Methods: DELETE, PATCH
	// Scopes: write // TODO: Add scopes
	// Description: Delete or update a token
	router.AddHandlerFuncRe(ctx, reToken, service.UpdateToken, http.MethodDelete, http.MethodPatch)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get all tokens
func (service *auth) ListTokens(w http.ResponseWriter, r *http.Request) {
	tokens := service.jar.Tokens()
	result := make([]*Token, 0, len(tokens))
	for _, token := range tokens {
		token.Value = ""
		result = append(result, &token)
	}
	httpresponse.JSON(w, result, http.StatusOK, jsonIndent)
}

// Get a token
func (service *auth) GetToken(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	token := service.jar.Get(urlParameters[0])
	if !token.IsValid() {
		httpresponse.Error(w, http.StatusNotFound)
		return
	}

	// Remove the token value before returning
	token.Value = ""

	// Return the token
	httpresponse.JSON(w, token, http.StatusOK, jsonIndent)
}

// Create a token
func (service *auth) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req TokenCreate

	// TODO: Parse Duration
	// TODO: Require unique name

	// Get the request
	if err := httprequest.Read(r, &req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create the token
	token := NewToken(req.Name, service.tokenBytes, req.Duration, req.Scope...)
	if !token.IsValid() {
		httpresponse.Error(w, http.StatusInternalServerError)
		return
	}

	// Add the token
	if err := service.jar.Create(token); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Remove the access_time which doesn't make sense when
	// creating a token
	token.Time = time.Time{}

	// Return the token
	httpresponse.JSON(w, token, http.StatusCreated, jsonIndent)
}

// Update an existing token
func (service *auth) UpdateToken(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	token := service.jar.Get(urlParameters[0])
	if !token.IsValid() {
		httpresponse.Error(w, http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := service.jar.Delete(token.Value); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		httpresponse.Error(w, http.StatusMethodNotAllowed)
		return
	}

	// Respond with no content
	httpresponse.Empty(w, http.StatusOK)
}
