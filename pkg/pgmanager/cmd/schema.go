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
	Schemas      SchemaListCommand   `cmd:"" group:"DATABASE" help:"List schemas"`
	Schema       SchemaGetCommand    `cmd:"" group:"DATABASE" help:"Get schema"`
	CreateSchema SchemaCreateCommand `cmd:"" group:"DATABASE" help:"Create a new schema"`
	UpdateSchema SchemaUpdateCommand `cmd:"" group:"DATABASE" help:"Update a schema"`
	DeleteSchema SchemaDeleteCommand `cmd:"" group:"DATABASE" help:"Delete a schema"`
}

type SchemaListCommand struct {
	schema.SchemaListRequest
}

type SchemaGetCommand struct {
	Name string `arg:"" name:"name" help:"Schema name ('database/schema')"`
}

type SchemaDeleteCommand struct {
	SchemaGetCommand
	Force bool `help:"Force delete"`
}

type SchemaCreateCommand struct {
	schema.SchemaMeta
	// Name  string   `arg:"" name:"name" help:"Database name"`
	// Owner string   `help:"Database owner"`
	// Acl   []string `help:"Access privileges"`
}

type SchemaUpdateCommand struct {
	Name string `help:"New schema name"`
	SchemaCreateCommand
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

func (cmd SchemaGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		schema, err := provider.GetSchema(ctx, cmd.Name)
		if err != nil {
			return err
		}

		// Print schema
		fmt.Println(schema)
		return nil
	})
}

func (cmd SchemaCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Perform request
		schema, err := provider.CreateSchema(ctx, cmd.SchemaMeta)
		if err != nil {
			return err
		}

		// Print schema
		fmt.Println(schema)
		return nil
	})
}

func (cmd SchemaDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteSchema(ctx, cmd.Name, client.WithForce(cmd.Force))
	})
}

func (cmd SchemaUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Swap names
		cmd.SchemaCreateCommand.Name, cmd.Name = cmd.Name, cmd.SchemaCreateCommand.Name

		// Perform request
		schema, err := provider.UpdateSchema(ctx, cmd.Name, cmd.SchemaMeta)
		if err != nil {
			return err
		}

		// Print schema
		fmt.Println(schema)
		return nil
	})
}
