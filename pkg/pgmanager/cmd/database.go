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

type DatabaseCommands struct {
	Databases      DatabaseListCommand   `cmd:"" group:"DATABASE" help:"List databases"`
	Database       DatabaseGetCommand    `cmd:"get" group:"DATABASE" help:"Get database"`
	CreateDatabase DatabaseCreateCommand `cmd:"create" group:"DATABASE" help:"Create a new database"`
	UpdateDatabase DatabaseUpdateCommand `cmd:"update" group:"DATABASE" help:"Update a database"`
	DeleteDatabase DatabaseDeleteCommand `cmd:"delete" group:"DATABASE" help:"Delete a database"`
}

type DatabaseListCommand struct {
	schema.DatabaseListRequest
}

type DatabaseGetCommand struct {
	Name string `arg:"" name:"name" help:"Database name"`
}

type DatabaseCreateCommand struct {
	schema.DatabaseMeta
	// Name  string   `arg:"" name:"name" help:"Database name"`
	// Owner string   `help:"Database owner"`
	// Acl   []string `help:"Access privileges"`
}

type DatabaseDeleteCommand struct {
	DatabaseGetCommand
	Force bool `help:"Force delete"`
}

type DatabaseUpdateCommand struct {
	Name string `help:"New database name"`
	DatabaseCreateCommand
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

func (cmd DatabaseGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		database, err := provider.GetDatabase(ctx, cmd.Name)
		if err != nil {
			return err
		}

		// Print database
		fmt.Println(database)
		return nil
	})
}

func (cmd DatabaseCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		database, err := provider.CreateDatabase(ctx, cmd.DatabaseMeta)
		if err != nil {
			return err
		}

		// Print database
		fmt.Println(database)
		return nil
	})
}

func (cmd DatabaseDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteDatabase(ctx, cmd.Name, client.WithForce(cmd.Force))
	})
}

func (cmd DatabaseUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Swap names
		cmd.DatabaseCreateCommand.Name, cmd.Name = cmd.Name, cmd.DatabaseCreateCommand.Name

		// Perform request
		database, err := provider.UpdateDatabase(ctx, cmd.Name, cmd.DatabaseCreateCommand.DatabaseMeta)
		if err != nil {
			return err
		}

		// Print database
		fmt.Println(database)
		return nil
	})
}
