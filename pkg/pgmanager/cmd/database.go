package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgmanager/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type DatabaseCommands struct {
	Databases DatabaseListCommand `cmd:"" group:"DATABASE" help:"List databases"`
}

type DatabaseListCommand struct {
	server.CmdOffsetLimit
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd DatabaseListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		databases, err := provider.ListDatabases(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print databases
		fmt.Println(databases)
		return nil
	})
}
