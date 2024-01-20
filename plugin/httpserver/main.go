package main

import (
	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
)

func Block() hcl.Block {
	return httpserver.Config{}
}
