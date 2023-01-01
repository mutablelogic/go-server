package main

import (
	// Modules
	mdns "github.com/mutablelogic/go-server/pkg/mdns"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return mdns.Plugin{}
}
