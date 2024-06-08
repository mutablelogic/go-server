package ldap

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

func (service *ldap) ScopeRead() []string {
	// Return read (list, get) scopes
	return []string{
		scopePrefix + service.Label() + "/read",
		scopePrefix + defaultName + "/read",
	}
}

func (service *ldap) ScopeWrite() []string {
	// Return write (create, delete) scopes
	return []string{
		scopePrefix + service.Label() + "/write",
		scopePrefix + defaultName + "/write",
	}
}
