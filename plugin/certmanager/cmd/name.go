package cmd

import (
	"context"
	"fmt"

	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/plugin/certmanager/client"
)

// Packages

// Plugins

///////////////////////////////////////////////////////////////////////////////
// TYPES

type NameCommands struct {
	Names NameListCommand `cmd:"" group:"CERTIFICATES" help:"List names"`
}

type NameListCommand struct{}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *NameListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// List names
		names, err := provider.ListNames(ctx)
		if err != nil {
			return err
		}

		// Print names
		fmt.Println(names)
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func run(ctx server.Cmd, fn func(context.Context, *client.Client) error) error {
	// Create a client
	provider, err := client.New(ctx.GetEndpoint("${CERT_PREFIX}").String(), ctx.GetClientOpts()...)
	if err != nil {
		return err
	}
	// Run the function
	return fn(ctx.Context(), provider)
}
