package auth

import (
	"context"
	"net/http"
	"strings"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Check interfaces are satisfied
var _ server.Middleware = (*auth)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultBearerHeader = "Authorization"
	defaultBearerType   = "Bearer"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (middleware *auth) Wrap(ctx context.Context, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenValue string

		// If bearer is true, get the token from the Authorization: Bearer header
		if middleware.bearer {
			tokenValue = strings.ToLower(strings.TrimSpace(getBearer(r)))
		}

		// Get token from request
		if tokenValue == "" {
			httpresponse.Error(w, http.StatusUnauthorized)
			return
		}

		// TODO: Hook for getting JWT from request here

		// Get token from the jar - check it is found and valid
		token := middleware.jar.GetWithValue(tokenValue)
		authorized := true
		if token.IsZero() {
			authorized = false
		} else if !token.IsValid() {
			authorized = false
		} else if token.IsScope(ScopeRoot) {
			// Allow
		} else {
			// TODO: Get scope for the route
			var allowedScopes = []string{}
			if token.IsScope(allowedScopes...) {
				authorized = true
			}
		}

		// Return unauthorized if token is not found or not valid
		if !authorized {
			httpresponse.Error(w, http.StatusUnauthorized)
			return
		}

		// TODO: Hook for setting JWT cookie here

		// Create a new context with the token name and scopes
		r = r.WithContext(WithToken(r.Context(), token))

		// Call next handler in chain
		next(w, r)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Get bearer token from request, or return empty string
func getBearer(r *http.Request) string {
	// Get the bearer token
	if value := r.Header.Get(defaultBearerHeader); value == "" {
		return ""
	} else if parts := strings.SplitN(value, " ", 2); len(parts) != 2 {
		return ""
	} else if parts[0] != defaultBearerType {
		return ""
	} else {
		return parts[1]
	}
}
