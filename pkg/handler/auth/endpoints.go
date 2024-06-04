package auth

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.ServiceEndpoints = (*auth)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	jsonIndent = 2
)

var (
	reRoot  = regexp.MustCompile(`^/?$`)
	reToken = regexp.MustCompile(`^/(` + types.ReIdentifier + `)/?$`)
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

	// Path: /<token-name>
	// Methods: GET
	// Scopes: read // TODO: Add scopes
	// Description: Get a token
	router.AddHandlerFuncRe(ctx, reToken, service.GetToken, http.MethodGet)

	// Path: /<token-name>
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
		// Remove the token value
		token.Value = ""
		result = append(result, &token)
	}
	httpresponse.JSON(w, result, http.StatusOK, jsonIndent)
}

// Get a token
func (service *auth) GetToken(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	token := service.jar.GetWithName(strings.ToLower(urlParameters[0]))
	if token.IsZero() {
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

	// Get the request
	if err := httprequest.Read(r, &req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check for a valid name
	req.Name = strings.TrimSpace(req.Name)
	if !types.IsIdentifier(req.Name) {
		httpresponse.Error(w, http.StatusBadRequest, "invalid 'name'")
		return
	} else if token := service.jar.GetWithName(req.Name); !token.IsZero() {
		httpresponse.Error(w, http.StatusConflict, "duplicate 'name'")
		return
	} else if duration := req.Duration.Duration; duration > 0 {
		// Truncate duration to minute, check
		duration = duration.Truncate(time.Minute)
		if duration < time.Minute {
			httpresponse.Error(w, http.StatusBadRequest, "invalid 'duration'")
			return
		} else {
			req.Duration.Duration = duration
		}
	}

	// TODO: Should not be able to add the root scope unless you have the root scope

	// Create the token
	token := NewToken(req.Name, service.tokenBytes, req.Duration.Duration, req.Scope...)
	if !token.IsValid() {
		httpresponse.Error(w, http.StatusInternalServerError)
		return
	}

	// Add the token to the jar
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

// Update (patch, delete) an existing token
func (service *auth) UpdateToken(w http.ResponseWriter, r *http.Request) {
	urlParameters := router.Params(r.Context())
	token := service.jar.GetWithName(strings.ToLower(urlParameters[0]))
	if token.IsZero() {
		httpresponse.Error(w, http.StatusNotFound)
		return
	}

	requestor := TokenName(r.Context())
	if requestor == token.Name {
		httpresponse.Error(w, http.StatusForbidden, "Cannot update the requestor's token")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := service.jar.Delete(token.Value); err != nil {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		// TODO: PATCH
		// Patch can be with name, expire_time, scopes
		httpresponse.Error(w, http.StatusMethodNotAllowed)
		return
	}

	// Respond with no content
	httpresponse.Empty(w, http.StatusOK)
}
