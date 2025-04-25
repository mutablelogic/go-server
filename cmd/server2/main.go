package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	// Packages
	kong "github.com/alecthomas/kong"
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	Run    ServiceRunCommand    `cmd:"" group:"SERVICE" help:"Run the service"`
	Config ServiceConfigCommand `cmd:"" group:"SERVICE" help:"Output configuration"`
}

type ServiceRunCommand struct {
	Plugins []string `flag:"" type:"path" help:"Plugins to load"`
}

type ServiceConfigCommand struct {
	Plugins []string `flag:"" type:"path" help:"Plugins to load"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	// Parse command-line flags
	var cli CLI
	kong := kong.Parse(&cli,
		kong.Name(execName()),
		kong.Description("command-line tool"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"HOST":        hostName(),
			"USER":        userName(),
			"CERT_PREFIX": "/cert/v1",
			"PG_PREFIX":   "/pg/v1",
			"AUTH_PREFIX": "/auth/v1",
		},
	)

	// Run
	//	kong.BindTo(app, (*server.Cmd)(nil))
	kong.FatalIfErrorf(kong.Run())
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ServiceRunCommand) Run() error {
	// Load plugins
	plugins, err := provider.LoadPluginsForPattern(cmd.Plugins...)
	if err != nil {
		return err
	}

	// Create a provider and resolve references
	provider, err := provider.New(func(ctx context.Context, label string, plugin server.Plugin) (server.Plugin, error) {
		fmt.Println(label)
		return plugin, nil
	}, plugins...)
	if err != nil {
		return err
	}

	// Run the server
	return provider.Run(context.Background())
}

func (cmd *ServiceConfigCommand) Run() error {
	// Load plugins
	plugins, err := provider.LoadPluginsForPattern(cmd.Plugins...)
	if err != nil {
		return err
	}

	// Print plugins
	for _, plugin := range plugins {
		meta, err := provider.New(plugin, plugin.Name())
		if err != nil {
			return err
		}
		fmt.Printf("\n// %s\n%s", plugin.Description(), meta)
	}

	// Run the server
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func hostName() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}

func userName() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.Username
}

func execName() string {
	// The name of the executable
	name, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Base(name)
}
