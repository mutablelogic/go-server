package main

import (
	"fmt"
	"os"

	// Packages
	auth "github.com/mutablelogic/go-server/pkg/auth/cmd"
	certmanager "github.com/mutablelogic/go-server/pkg/cert/cmd"
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	ldap "github.com/mutablelogic/go-server/pkg/ldap/cmd"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager/cmd"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	ServiceCommands

	PG struct {
		pgmanager.DatabaseCommands
		pgmanager.SchemaCommands
		pgmanager.ObjectCommands
		pgmanager.RoleCommands
		pgmanager.ConnectionCommands
		pgmanager.TablespaceCommands
		pgqueue.QueueCommands
		pgqueue.TaskCommands
		pgqueue.TickerCommands
	} `cmd:""`

	Cert struct {
		certmanager.NameCommands
		certmanager.CertCommands
	} `cmd:""`

	Auth struct {
		auth.UserCommands
		auth.TokenCommands
		auth.AuthCommands
	} `cmd:""`

	LDAP struct {
		ldap.ObjectCommands
		ldap.AuthCommands
	} `cmd:""`

	VersionCommands
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	// Parse command-line flags
	var cli CLI

	app, err := cmd.New(&cli, "go-server command-line tool")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
}
