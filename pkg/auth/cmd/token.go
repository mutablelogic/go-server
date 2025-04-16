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

type TokenCommands struct {
	Tokens      TokenListCommand   `cmd:"" group:"AUTH" help:"List tokens for user"`
	Token       TokenGetCommand    `cmd:"" group:"AUTH" help:"Get token for user"`
	CreateToken TokenCreateCommand `cmd:"" group:"AUTH" help:"Create new token for user"`
	UpdateToken TokenUpdateCommand `cmd:"" group:"AUTH" help:"Update token for user"`
	DeleteToken TokenDeleteCommand `cmd:"" group:"AUTH" help:"Delete token for user"`
}

type TokenListCommand struct {
	User string `arg:"" help:"User name"`
	schema.TokenListRequest
}

type TokenCreateCommand struct {
	User string `arg:"" help:"User name"`
	schema.TokenMeta
}

type TokenGetCommand struct {
	schema.TokenId
}

type TokenDeleteCommand struct {
	schema.TokenId
	Force bool `help:"Force delete"`
}

type TokenUpdateCommand struct {
	schema.TokenId
	schema.TokenMeta
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd TokenListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		tokens, err := provider.ListTokens(ctx, cmd.User, client.WithOffsetLimit(cmd.Offset, cmd.Limit), client.WithStatus(cmd.Status))
		if err != nil {
			return err
		}

		// List tokens
		fmt.Println(tokens)
		return nil
	})
}

func (cmd TokenCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		token, err := provider.CreateToken(ctx, cmd.User, cmd.TokenMeta)
		if err != nil {
			return err
		}

		// Print token
		fmt.Println(token)
		return nil
	})
}

func (cmd TokenGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		token, err := provider.GetToken(ctx, cmd.User, cmd.Id)
		if err != nil {
			return err
		}

		// Print token
		fmt.Println(token)
		return nil
	})
}

func (cmd TokenDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteToken(ctx, cmd.User, cmd.Id, cmd.Force)
	})
}

func (cmd TokenUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		token, err := provider.UpdateToken(ctx, cmd.User, cmd.Id, cmd.TokenMeta)
		if err != nil {
			return err
		}

		// Print token
		fmt.Println(token)
		return nil
	})
}
