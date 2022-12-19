package main

import (
	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver/router"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return router.Plugin{}
}
