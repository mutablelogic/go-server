package plugin

import (
	"time"
)

// TokenAuth is an interface for access to API methods through tokens
type TokenAuth interface {
	// Return true if a token associated with the name already exists
	Exists(string) bool

	// Create a new token associated with a name, duration and scopes.
	// Return the token value. The duration can be zero for no expiry.
	Create(string, time.Duration, ...string) (string, error)

	// Revoke a token associated with a name. For the admin token, it is
	// rotated rather than revoked.
	Revoke(string) error

	// Return all token names and their last access times, including
	// expired tokens
	Enumerate() map[string]time.Time

	// Returns the name of the token if a value matches and is
	// valid. Updates the access time for the token. If token with value not
	// found, then return empty string
	Matches(string) string

	// Returns true if the named token is valid, and the scope matches.
	MatchesScope(name, scope string) bool
}
