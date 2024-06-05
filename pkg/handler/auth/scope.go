package auth

import (
	// Packages
	"github.com/mutablelogic/go-server/pkg/version"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Prefix
	scopePrefix = version.GitSource + "/scope/"

	// Root scope allows ANY operation
	ScopeRoot = scopePrefix + "root"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (auth *auth) ScopeRead() []string {
	// Return read (list, get) scopes
	return []string{
		scopePrefix + auth.Label() + "/read",
		scopePrefix + defaultName + "/read",
	}
}

func (auth *auth) ScopeWrite() []string {
	// Return write (create, delete, update) scopes
	return []string{
		scopePrefix + auth.Label() + "/write",
		scopePrefix + defaultName + "/write",
	}
}
