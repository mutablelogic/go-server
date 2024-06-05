package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	logger "github.com/mutablelogic/go-server/pkg/middleware/logger"
)

func Plugin() server.Plugin {
	return logger.Config{}
}
