package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	logger "github.com/mutablelogic/go-server/pkg/handler/logger"
)

func Plugin() server.Plugin {
	return logger.Config{}
}
