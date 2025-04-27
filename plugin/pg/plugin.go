package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager/config"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Plugin() server.Plugin {
	return pgmanager.Config{}
}
