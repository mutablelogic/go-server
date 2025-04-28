package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgqueue/client"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TickerCommands struct {
	Tickers      TickerListCommand   `cmd:"" group:"QUEUE" help:"List tickers"`
	NextTicker   TickerNextCommand   `cmd:"next" group:"QUEUE" help:"Receive ticker events"`
	Ticker       TickerGetCommand    `cmd:"list" group:"QUEUE" help:"Get ticker by name"`
	CreateTicker TickerCreateCommand `cmd:"create" group:"QUEUE" help:"Create a new ticker"`
	UpdateTicker TickerUpdateCommand `cmd:"update" group:"QUEUE" help:"Update ticker"`
	DeleteTicker TickerDeleteCommand `cmd:"delete" group:"QUEUE" help:"Delete ticker"`
}

type TickerListCommand struct {
	schema.TickerListRequest
}

type TickerCreateCommand struct {
	Ticker   string         `arg:"" help:"Ticker name"`
	Payload  *string        `help:"JSON Ticker payload"`
	Interval *time.Duration `help:"Interval (default 1 minute)"`
}

type TickerGetCommand struct {
	Ticker string `arg:"" help:"Ticker name"`
}

type TickerUpdateCommand struct {
	Ticker string `help:"New ticker name"`
	TickerCreateCommand
}

type TickerDeleteCommand struct {
	TickerGetCommand
}

type TickerNextCommand struct {
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

func (cmd TickerGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		ticker, err := provider.GetTicker(ctx, cmd.Ticker)
		if err != nil {
			return err
		}

		// Print ticker
		fmt.Println(ticker)
		return nil
	})
}

func (cmd TickerCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Get ticker metadata
		meta := schema.TickerMeta{
			Ticker: cmd.Ticker,
		}
		if cmd.Payload != nil {
			if err := json.Unmarshal([]byte(types.PtrString(cmd.Payload)), &meta.Payload); err != nil {
				return fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}
		if cmd.Interval != nil {
			meta.Interval = types.DurationPtr(*cmd.Interval)
		}

		// Create ticker
		ticker, err := provider.CreateTicker(ctx, meta)
		if err != nil {
			return err
		}

		// Print ticker
		fmt.Println(ticker)
		return nil
	})
}

func (cmd TickerDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteTicker(ctx, cmd.Ticker)
	})
}

func (cmd TickerUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Swap ticker names
		cmd.Ticker, cmd.TickerCreateCommand.Ticker = cmd.TickerCreateCommand.Ticker, cmd.Ticker

		// Set the metadata
		meta := schema.TickerMeta{
			Ticker:   cmd.Ticker,
			Interval: cmd.Interval,
		}
		if cmd.Payload != nil {
			if err := json.Unmarshal([]byte(types.PtrString(cmd.Payload)), &meta.Payload); err != nil {
				return fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		// Update the ticker
		ticker, err := provider.UpdateTicker(ctx, cmd.Ticker, meta)
		if err != nil {
			return err
		}

		// Print ticker
		fmt.Println(ticker)
		return nil
	})
}

func (cmd TickerNextCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.NextTicker(ctx)
	})
}
