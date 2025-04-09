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

type SchemaCommands struct {
	Schemas SchemaListCommand `cmd:"" group:"DATABASE" help:"List schemas"`
}

type SchemaListCommand struct {
	schema.SchemaListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd SchemaListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		databases, err := provider.ListSchemas(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print databases
		fmt.Println(databases)
		return nil
	})
}
