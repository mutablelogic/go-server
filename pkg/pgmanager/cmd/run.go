package cmd

import (
	"context"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgmanager/client"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func run(ctx server.Cmd, fn func(context.Context, *client.Client) error) error {
	// Create a client
	provider, err := client.New(ctx.GetEndpoint("${PG_PREFIX}").String(), ctx.GetClientOpts()...)
	if err != nil {
		return err
	}
	// Run the function
	return fn(ctx.Context(), provider)
}
