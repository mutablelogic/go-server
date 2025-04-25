package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	httpserver "github.com/mutablelogic/go-server/plugin/httpserver/config"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Plugin() server.Plugin {
	return httpserver.Config{}
}
