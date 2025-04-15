package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/auth/client"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type UserCommands struct {
	Users      UserListCommand   `cmd:"" group:"AUTH" help:"List users"`
	CreateUser UserCreateCommand `cmd:"" group:"AUTH" help:"Create a new user"`
}

type UserListCommand struct {
	schema.UserListRequest
}

type UserCreateCommand struct {
	schema.UserMeta
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd UserListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		users, err := provider.ListUsers(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit), client.WithStatus(cmd.Status), client.WithScope(cmd.Scope))
		if err != nil {
			return err
		}

		// List users
		fmt.Println(users)
		return nil
	})
}

func (cmd UserCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// If the user is an email address, then set the descrption field
		var name, email string
		if types.IsEmail(types.PtrString(cmd.UserMeta.Name), &name, &email) {
			cmd.UserMeta.Name = types.StringPtr(email)
			if cmd.UserMeta.Desc == nil {
				cmd.UserMeta.Desc = types.StringPtr(name)
			}
		}

		// Create the user
		user, err := provider.CreateUser(ctx, cmd.UserMeta)
		if err != nil {
			return err
		}

		// Print user
		fmt.Println(user)
		return nil
	})
}
