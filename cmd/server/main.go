package main

import (
	"os"
	"os/user"
	"path/filepath"

	// Packages
	kong "github.com/alecthomas/kong"
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/auth/cmd"
	auth_schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	certmanager "github.com/mutablelogic/go-server/pkg/cert/cmd"
	cert_schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager/cmd"
	pgmanager_schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue/cmd"
	pgqueue_schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
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
	pgmanager.TablespaceCommands
	pgqueue.QueueCommands
	pgqueue.TaskCommands
	pgqueue.TickerCommands
	certmanager.NameCommands
	certmanager.CertCommands
	auth.UserCommands
	auth.TokenCommands
	auth.AuthCommands
	VersionCommands
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
			"HOST":         hostName(),
			"USER":         userName(),
			"CERT_PREFIX":  cert_schema.APIPrefix,
			"PG_PREFIX":    pgmanager_schema.APIPrefix,
			"AUTH_PREFIX":  auth_schema.APIPrefix,
			"QUEUE_PREFIX": pgqueue_schema.APIPrefix,
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
