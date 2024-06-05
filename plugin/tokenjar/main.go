package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	tokenjar "github.com/mutablelogic/go-server/pkg/handler/tokenjar"
)

func Plugin() server.Plugin {
	return tokenjar.Config{}
}
