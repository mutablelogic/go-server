package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth"
)

func Plugin() server.Plugin {
	return auth.Config{}
}
