package main

import (
	// Package imports
	log "github.com/mutablelogic/go-server/pkg/log"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

// BuiltInPlugins is a list of plugins which are compiled into the binary
var BuiltInPlugins = []Plugin{
	log.Plugin{},
}

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

// Add log plugin if one does not exist
func AddLogPlugin(plugins task.Plugins) error {
	if plugin := plugins.Get(log.DefaultName); plugin == nil {
		return plugins.Register(log.WithLabel("main"))
	}
	return nil
}
