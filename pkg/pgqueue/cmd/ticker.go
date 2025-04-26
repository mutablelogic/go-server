package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgqueue/client"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TickerCommands struct {
	Tickers      TickerListCommand   `cmd:"" group:"QUEUE" help:"List tickers"`
	CreateTicker TickerCreateCommand `cmd:"create" group:"QUEUE" help:"Create a new ticker"`
}

type TickerListCommand struct {
	schema.TickerListRequest
}

type TickerCreateCommand struct {
	schema.TickerMeta
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd TickerListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		tickers, err := provider.ListTickers(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print tickers
		fmt.Println(tickers)
		return nil
	})
}

func (cmd TickerCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		ticker, err := provider.CreateTicker(ctx, cmd.TickerMeta)
		if err != nil {
			return err
		}

		// Print database
		fmt.Println(ticker)
		return nil
	})
}
