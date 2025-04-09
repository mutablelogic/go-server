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
	server.CmdOffsetLimit
}

type DatabaseGetCommand struct {
	Name string `arg:"" name:"name" help:"Database name"`
}

type DatabaseCreateCommand struct {
	Name  string   `arg:"" name:"name" help:"Database name"`
	Owner string   `help:"Database owner"`
	Acl   []string `help:"Access privileges"`
}

type DatabaseDeleteCommand struct {
	DatabaseGetCommand
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
		acl, err := schema.ParseACL(cmd.Acl)
		if err != nil {
			return err
		}
		database, err := provider.CreateDatabase(ctx, schema.Database{
			Name:  cmd.Name,
			Owner: cmd.Owner,
			Acl:   acl,
		})
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
		return provider.DeleteDatabase(ctx, cmd.Name)
	})
}

func (cmd DatabaseUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Swap names
		cmd.DatabaseCreateCommand.Name, cmd.Name = cmd.Name, cmd.DatabaseCreateCommand.Name

		// Parse ACL's
		acl, err := schema.ParseACL(cmd.Acl)
		if err != nil {
			return err
		}

		// Perform request
		database, err := provider.UpdateDatabase(ctx, cmd.Name, schema.Database{
			Name:  cmd.DatabaseCreateCommand.Name,
			Owner: cmd.Owner,
			Acl:   acl,
		})
		if err != nil {
			return err
		}

		// Print database
		fmt.Println(database)
		return nil
	})
}
