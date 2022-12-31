package main

import (
	// Packages
	"github.com/mutablelogic/go-server/npm/router"
	"github.com/mutablelogic/go-server/pkg/httpserver/static"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

var _ = router.Dist

func Config() Plugin {
	return static.Plugin{}
}
