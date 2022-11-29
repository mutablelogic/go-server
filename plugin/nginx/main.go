package main

import (
	// Modules
	nginx "github.com/mutablelogic/go-server/pkg/nginx"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return nginx.Plugin{}
}
