package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	metrics "github.com/mutablelogic/go-server/pkg/metrics/config"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Plugin() server.Plugin {
	return metrics.Config{}
}
