package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/ldap/client"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type UserCommands struct {
	Users      UserListCommand   `cmd:"" group:"LDAP" help:"List users"`
	User       UserGetCommand    `cmd:"" group:"LDAP" help:"Get user"`
	CreateUser UserCreateCommand `cmd:"" group:"LDAP" help:"Create user"`
	DeleteUser UserDeleteCommand `cmd:"" group:"LDAP" help:"Delete user"`
}

type UserListCommand struct {
	schema.ObjectListRequest
}

type UserGetCommand struct {
	User string `arg:"" help:"Username"`
}

type UserCreateCommand struct {
	UserGetCommand
	Attr []string `arg:"" help:"attribute=value,value,..."`
}

type UserDeleteCommand struct {
	UserGetCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd UserListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		users, err := provider.ListUsers(ctx, client.WithFilter(cmd.Filter), client.WithAttr(cmd.Attr...), client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print users
		fmt.Println(users)
		return nil
	})
}

func (cmd UserGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		user, err := provider.GetUser(ctx, cmd.User)
		if err != nil {
			return err
		}

		// Print user
		fmt.Println(user)
		return nil
	})
}

func (cmd UserDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteUser(ctx, cmd.User)
	})
}

func (cmd UserCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Decode attributes
		attrs := url.Values{}
		for _, attr := range cmd.Attr {
			parts := strings.SplitN(attr, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid attribute: %s", attr)
			}
			name := parts[0]
			if values := strings.Split(parts[1], ","); len(values) > 0 {
				attrs[name] = values
			}
		}

		// Create user
		user, err := provider.CreateUser(ctx, schema.Object{
			DN:     cmd.User,
			Values: attrs,
		})
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(user)
		return nil
	})
}
