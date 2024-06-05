package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	static "github.com/mutablelogic/go-server/pkg/handler/static"
)

func Plugin() server.Plugin {
	return static.Config{}
}
