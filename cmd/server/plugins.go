package main

import (
	// Package imports
	dnsregister "github.com/mutablelogic/go-server/pkg/dnsregister"
	log "github.com/mutablelogic/go-server/pkg/log"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

// BuiltInPlugins is a list of plugins which are compiled into the binary
var BuiltInPlugins = []Plugin{
	log.Plugin{},
	dnsregister.Plugin{},
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
