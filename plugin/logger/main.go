package main

import (
	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	logger "github.com/mutablelogic/go-server/pkg/logger"
)

func Block() hcl.Block {
	return logger.Config{}
}
