package main

import (
	logger "github.com/mutablelogic/go-server/pkg/logger"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

var BuiltInPlugins = []Plugin{logger.Plugin{}}

// BuiltinPlugins returns the list of plugins which are compiled into the binary
func BuiltinPlugins() (task.Plugins, error) {
	builtins := task.Plugins{}
	for _, p := range BuiltInPlugins {
		if err := builtins.Register(p); err != nil {
			return nil, err
		}
	}
	return builtins, nil
}
