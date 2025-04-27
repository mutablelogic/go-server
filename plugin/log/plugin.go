package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	log "github.com/mutablelogic/go-server/pkg/logger/config"
)

func Plugin() server.Plugin {
	return log.Config{}
}
