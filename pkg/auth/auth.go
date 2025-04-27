package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ref "github.com/mutablelogic/go-server/pkg/ref"
)

/////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ApiKeyHeader = "X-API-Key"
	ApiKeyQuery  = "api-key"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns the API token from the request, checking for the X-API-Key header first
// and then the api-key query parameter (not recommended)
func RequireScope(w http.ResponseWriter, r *http.Request, fn func(w http.ResponseWriter, r *http.Request) error, scopes ...string) error {
	// Get the authorization method from the context
	auth := ref.Auth(r.Context())
	if auth == nil {
		// No auth provider, proceed without authentication
		return fn(w, r)
	}

	// Authenticate the user
	user, err := auth.Authenticate(r)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Authorize the user  for the given scopes
	if err := auth.Authorize(r.Context(), user, scopes...); err != nil {
		return httpresponse.Error(w, err)
	}

	// Add the authenticated user to the request context
	ctx := ref.WithUser(r.Context(), user)

	// Call the function with the user
	return fn(w, r.WithContext(ctx))
}

// Returns token from the request, checking for the X-API-Key header first
func GetRequestToken(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get(ApiKeyHeader)); token != "" {
		return token
	}
	if token := strings.TrimSpace(r.URL.Query().Get(ApiKeyQuery)); token != "" {
		return token
	}
	return ""
}

/////////////////////////////////////////////////////////////////////////////////
// AUTHZ

// Return a "live" user for a HTTP request
func (manager *Manager) Authenticate(r *http.Request) (*schema.User, error) {
	token := GetRequestToken(r)
	if token == "" {
		return nil, httpresponse.Err(http.StatusUnauthorized)
	}

	// Check user
	user, err := manager.GetUserForToken(r.Context(), token)
	if errors.Is(err, httpresponse.ErrNotFound) {
		return nil, httpresponse.Err(http.StatusUnauthorized)
	} else if err != nil {
		return nil, err
	} else if schema.UserStatus(user.Status) != schema.UserStatusLive {
		return nil, httpresponse.Err(http.StatusUnauthorized)
	}

	// Return success
	return user, nil
}

// Authorize the user against scopes
func (manager *Manager) Authorize(ctx context.Context, user *schema.User, scopes ...string) error {
	// TODO
	return nil
}
