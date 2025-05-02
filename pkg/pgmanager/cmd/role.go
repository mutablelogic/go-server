package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgmanager/client"
	"github.com/mutablelogic/go-server/pkg/pgmanager/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type RoleCommands struct {
	Roles      RoleListCommand   `cmd:"" group:"DATABASE" help:"List database roles"`
	Role       RoleGetCommand    `cmd:"get" group:"DATABASE" help:"Get a database role"`
	CreateRole RoleCreateCommand `cmd:"create" group:"DATABASE" help:"Create a new database role"`
	UpdateRole RoleUpdateCommand `cmd:"update" group:"DATABASE" help:"Update a database role"`
	DeleteRole RoleDeleteCommand `cmd:"delete" group:"DATABASE" help:"Delete a database role"`
}

type RoleListCommand struct {
	schema.RoleListRequest
}

type RoleCreateCommand struct {
	schema.RoleMeta
}

type RoleGetCommand struct {
	Name string `arg:"" name:"name" help:"Role name"`
}

type RoleDeleteCommand struct {
	RoleGetCommand
}

type RoleUpdateCommand struct {
	Name string `help:"New role name"`
	RoleCreateCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd RoleListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		names, err := provider.ListRoles(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print roles
		fmt.Println(names)
		return nil
	})
}

func (cmd RoleCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		role, err := provider.CreateRole(ctx, cmd.RoleMeta)
		if err != nil {
			return err
		}

		// Print role
		fmt.Println(role)
		return nil
	})
}

func (cmd RoleGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		name, err := provider.GetRole(ctx, cmd.Name)
		if err != nil {
			return err
		}

		// Print role
		fmt.Println(name)
		return nil
	})
}

func (cmd RoleDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteRole(ctx, cmd.Name)
	})
}

func (cmd RoleUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Swap names
		cmd.RoleCreateCommand.Name, cmd.Name = cmd.Name, cmd.RoleCreateCommand.Name

		// Perform request
		role, err := provider.UpdateRole(ctx, cmd.Name, cmd.RoleCreateCommand.RoleMeta)
		if err != nil {
			return err
		}

		// Print role
		fmt.Println(role)
		return nil
	})
}
