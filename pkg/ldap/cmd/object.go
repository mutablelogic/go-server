package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/ldap/client"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ObjectCommands struct {
	Objects ObjectListCommand `cmd:"" group:"LDAP" help:"List queues"`
}

type ObjectListCommand struct {
	schema.ObjectListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd ObjectListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		queues, err := provider.ListObjects(ctx, client.WithFilter(cmd.Filter), client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print queues
		fmt.Println(queues)
		return nil
	})
}
