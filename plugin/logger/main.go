package main

import (
	// Packages
	hcl "github.com/mutablelogic/go-server/pkg/hcl"
	logger "github.com/mutablelogic/go-server/pkg/logger"
)

func Block() hcl.Block {
	return logger.Config{}
}
