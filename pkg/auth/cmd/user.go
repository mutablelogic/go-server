package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/auth/client"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type UserCommands struct {
	Users UserListCommand `cmd:"" group:"AUTH" help:"List users"`
}

type UserListCommand struct {
	schema.UserListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd UserListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		users, err := provider.ListUsers(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit), client.WithStatus(cmd.Status), client.WithScope(cmd.Scope))
		if err != nil {
			return err
		}

		// Print databases
		fmt.Println(users)
		return nil
	})
}
