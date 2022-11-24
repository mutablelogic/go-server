package main

import (
	// Modules
	dnsregister "github.com/mutablelogic/go-server/pkg/dnsregister"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Config() Plugin {
	return dnsregister.Plugin{}
}
