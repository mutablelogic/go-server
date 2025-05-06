package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/ldap/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type AuthCommands struct {
	Auth           AuthBindCommand           `cmd:"" group:"LDAP" help:"Authenticate user"`
	ChangePassword AuthChangePasswordCommand `cmd:"" group:"LDAP" help:"Change password"`
}

type AuthBindCommand struct {
	ObjectGetCommand
	Password string `required:"" help:"Password"`
}

type AuthChangePasswordCommand struct {
	AuthBindCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd AuthBindCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		object, err := provider.Auth(ctx, cmd.DN, cmd.Password)
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(object)
		return nil
	})
}

func (cmd AuthChangePasswordCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		object, err := provider.ChangePassword(ctx, cmd.DN, cmd.Password)
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(object)
		return nil
	})
}
