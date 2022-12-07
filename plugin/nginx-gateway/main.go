package main

import (
	// Modules
	gateway "github.com/mutablelogic/go-server/pkg/nginx/gateway"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return gateway.Plugin{}
}
