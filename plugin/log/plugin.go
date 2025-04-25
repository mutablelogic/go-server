package main

import (
	server "github.com/mutablelogic/go-server"
	log "github.com/mutablelogic/go-server/plugin/log/config"
)

func Plugin() server.Plugin {
	return log.Config{}
}
