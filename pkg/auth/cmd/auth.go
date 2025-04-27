package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/auth/client"
)

// Packages

///////////////////////////////////////////////////////////////////////////////
// TYPES

type AuthCommands struct {
	Auth AuthCommand `cmd:"" group:"AUTH" help:"Return a user from an auth token"`
}

type AuthCommand struct {
	Value string `arg:"" help:"Token value"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd AuthCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		user, err := provider.Auth(ctx, cmd.Value)
		if err != nil {
			return err
		}

		// Print user
		fmt.Println(user)
		return nil
	})
}
