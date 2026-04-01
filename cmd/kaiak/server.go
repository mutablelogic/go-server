//go:build !client

package main

import (
	// Packages
	server "github.com/mutablelogic/go-server"
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run the server."`
	cmd.OpenAPICommands
}

type RunServer struct {
	cmd.RunServer

	// Static file serving options
	Static struct {
		Path string `name:"path" help:"URL path for static files (relative to prefix)" default:""`
		Dir  string `name:"dir" help:"Directory on disk to serve as static files" default:""`
	} `embed:"" prefix:"static."`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (s *RunServer) Run(ctx server.Cmd) error {
	return s.RunServer.Run(ctx)
}
