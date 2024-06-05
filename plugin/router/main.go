package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
)

func Plugin() server.Plugin {
	return router.Config{}
}
