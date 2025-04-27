package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgmanager/client"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ConnectionCommands struct {
	Connections    ConnectionListCommand   `cmd:"" group:"DATABASE" help:"List connections"`
	Connection     ConnectionGetCommand    `cmd:"get" group:"DATABASE" help:"Get connection"`
	KillConnection ConnectionDeleteCommand `cmd:"delete" group:"DATABASE" help:"Kill connection"`
}

type ConnectionListCommand struct {
	schema.ConnectionListRequest
}

type ConnectionGetCommand struct {
	Pid uint64 `arg:"" name:"pid" help:"Connection PID"`
}

type ConnectionDeleteCommand struct {
	ConnectionGetCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd ConnectionListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		connections, err := provider.ListConnections(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit), client.WithDatabase(cmd.Database), client.WithRole(cmd.Role), client.WithState(cmd.State))
		if err != nil {
			return err
		}

		// Print connections
		fmt.Println(connections)
		return nil
	})
}

func (cmd ConnectionGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		connection, err := provider.GetConnection(ctx, cmd.Pid)
		if err != nil {
			return err
		}

		// Print connection
		fmt.Println(connection)
		return nil
	})
}

func (cmd ConnectionDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteConnection(ctx, cmd.Pid)
	})
}
