package main

import (
	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	simplerouter "github.com/mutablelogic/go-server/pkg/simplerouter"
)

func Block() hcl.Block {
	return simplerouter.Config{}
}
