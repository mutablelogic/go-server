package main

import (
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/parser"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Service2Commands struct {
	//	Run  ServiceRunCommand  `cmd:"" group:"SERVICE" help:"Run the service"`
	Run2 Service2RunCommand `cmd:"" group:"SERVICE" help:"Run the service with plugins"`
}

type Service2RunCommand struct {
	Plugins []string `help:"Plugin paths" env:"PLUGIN_PATH"`
	Args    []string `arg:"" help:"Configuration files"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *Service2RunCommand) Run(app server.Cmd) error {
	// Create a provider by loading the plugins
	provider, err := provider.NewWithPlugins(cmd.Plugins...)
	if err != nil {
		return err
	}

	// Create a parser of the config files
	parser, err := parser.New(provider.Plugins()...)
	if err != nil {
		return err
	}

	// Parse the configuration files
	for _, path := range cmd.Args {
		if err := parser.Parse(path); err != nil {
			return err
		}
	}

	fmt.Println(provider, parser)

	return nil
}
