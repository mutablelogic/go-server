package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
)

func Plugin() server.Plugin {
	return httpserver.Config{}
}
