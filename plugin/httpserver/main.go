package main

import (
	// Modules
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return httpserver.Plugin{}
}
