package main

import (
	"os"
	"os/user"
	"path/filepath"

	// Packages
	kong "github.com/alecthomas/kong"
	server "github.com/mutablelogic/go-server"
	certmanager "github.com/mutablelogic/go-server/pkg/cert/cmd"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	Globals
	ServiceCommands
	pgmanager.DatabaseCommands
	pgmanager.SchemaCommands
	pgmanager.ObjectCommands
	pgmanager.RoleCommands
	pgmanager.ConnectionCommands
	certmanager.NameCommands
	certmanager.CertCommands
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

	// Create the app
	app, err := NewApp(cli.Globals, kong.Model.Vars())
	if err != nil {
		kong.FatalIfErrorf(err)
	}
	defer app.Close()

	// Run
	kong.BindTo(app, (*server.Cmd)(nil))
	kong.FatalIfErrorf(kong.Run())
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
