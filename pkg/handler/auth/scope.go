package auth

import "github.com/mutablelogic/go-server/pkg/version"

var (
	// Root scope allows ANY operation
	ScopeRoot = version.GitSource + "/scope/root"
)
