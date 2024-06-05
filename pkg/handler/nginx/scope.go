package nginx

import (
	// Packages
	"github.com/mutablelogic/go-server/pkg/version"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Prefix
	scopePrefix = version.GitSource + "/scope/"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (nginx *nginx) ScopeRead() []string {
	// Return read (list, get) scopes
	return []string{
		scopePrefix + nginx.Label() + "/read",
		scopePrefix + defaultName + "/read",
	}
}

func (nginx *nginx) ScopeWrite() []string {
	// Return write (create, delete, update) scopes
	return []string{
		scopePrefix + nginx.Label() + "/write",
		scopePrefix + defaultName + "/write",
	}
}
