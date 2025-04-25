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

type ObjectCommands struct {
	Objects ObjectListCommand `cmd:"" group:"DATABASE" help:"List objects"`
	Object  ObjectGetCommand  `cmd:"" group:"DATABASE" help:"Get object"`
}

type ObjectListCommand struct {
	schema.ObjectListRequest
}

type ObjectGetCommand struct {
	Database string
	schema.ObjectName
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd ObjectListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		databases, err := provider.ListObjects(ctx, client.WithDatabase(cmd.Database), client.WithSchema(cmd.Schema), client.WithType(cmd.Type), client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print databases
		fmt.Println(databases)
		return nil
	})
}

func (cmd ObjectGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		object, err := provider.GetObject(ctx, cmd.Database, cmd.Schema, cmd.Name)
		if err != nil {
			return err
		}

		// Print databases
		fmt.Println(object)
		return nil
	})
}
