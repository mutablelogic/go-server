package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	helloworld "github.com/mutablelogic/go-server/npm/helloworld"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Plugin() server.Plugin {
	return helloworld.Config{}
}
