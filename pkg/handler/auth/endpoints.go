package auth

import (
	"context"
	"net/http"
	"regexp"
	"slices"
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
func (service *auth) AddEndpoints(ctx context.Context, r server.Router) {
	// Path: /
	// Methods: GET
	// Scopes: read
	// Description: Get current set of tokens and groups
	r.AddHandlerFuncRe(ctx, reRoot, service.ListTokens, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /
	// Methods: POST
	// Scopes: write
	// Description: Create a new token
	r.AddHandlerFuncRe(ctx, reRoot, service.CreateToken, http.MethodPost).(router.Route).
		SetScope(service.ScopeWrite()...)

	// Path: /<token-name>
	// Methods: GET
	// Scopes: read
	// Description: Get a token
	r.AddHandlerFuncRe(ctx, reToken, service.GetToken, http.MethodGet).(router.Route).
		SetScope(service.ScopeRead()...)

	// Path: /<token-name>
	// Methods: DELETE, PATCH
	// Scopes: write
	// Description: Delete or update a token
	r.AddHandlerFuncRe(ctx, reToken, service.UpdateToken, http.MethodDelete, http.MethodPatch).(router.Route).
		SetScope(service.ScopeWrite()...)

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
	token := service.jar.GetWithName(urlParameters[0])
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
	if err := httprequest.Body(&req, r); err != nil {
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

	// Should not be able to add the root scope unless you have the root scope
	if slices.Contains(req.Scope, ScopeRoot) {
		requestorScopes := TokenScope(r.Context())
		if !slices.Contains(requestorScopes, ScopeRoot) {
			httpresponse.Error(w, http.StatusForbidden, "Cannot create a token with root scope")
		}
	}

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
	token := service.jar.GetWithName(urlParameters[0])
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
		// TODO: Should not be able to add the root scope unless you have the root scope
		httpresponse.Error(w, http.StatusMethodNotAllowed)
		return
	}

	// Respond with no content
	httpresponse.Empty(w, http.StatusOK)
}
