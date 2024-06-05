package router

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

func (router *router) ScopeRead() []string {
	// Return read (list, get) scopes
	return []string{
		scopePrefix + router.Label() + "/read",
		scopePrefix + defaultName + "/read",
	}
}

func (router *router) ScopeWrite() []string {
	// Return write (create, delete, update) scopes
	return []string{
		scopePrefix + router.Label() + "/write",
		scopePrefix + defaultName + "/write",
	}
}
