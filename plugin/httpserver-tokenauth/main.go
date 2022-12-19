package main

import (
	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver/tokenauth"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return tokenauth.Plugin{}
}
