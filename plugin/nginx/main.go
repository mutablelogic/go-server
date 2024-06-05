package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx"
)

func Plugin() server.Plugin {
	return nginx.Config{}
}
