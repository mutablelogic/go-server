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

type QueueCommands struct {
	Queues      QueueListCommand   `cmd:"" group:"QUEUE" help:"List queues"`
	Queue       QueueGetCommand    `cmd:"list" group:"QUEUE" help:"Get queue by name"`
	CreateQueue QueueCreateCommand `cmd:"create" group:"QUEUE" help:"Create a new queue"`
	UpdateQueue QueueUpdateCommand `cmd:"update" group:"QUEUE" help:"Update queue"`
	DeleteQueue QueueDeleteCommand `cmd:"delete" group:"QUEUE" help:"Delete queue"`
}

type QueueListCommand struct {
	schema.QueueListRequest
}

type QueueCreateCommand struct {
	schema.QueueMeta
}

type QueueGetCommand struct {
	Queue string `arg:"" help:"Queue name"`
}

type QueueUpdateCommand struct {
	Name string `help:"New Queue name"`
	schema.QueueMeta
}

type QueueDeleteCommand struct {
	QueueGetCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd QueueListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		queues, err := provider.ListQueues(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print queues
		fmt.Println(queues)
		return nil
	})
}

func (cmd QueueGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		queue, err := provider.GetQueue(ctx, cmd.Queue)
		if err != nil {
			return err
		}

		// Print queue
		fmt.Println(queue)
		return nil
	})
}

func (cmd QueueCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		queue, err := provider.CreateQueue(ctx, cmd.QueueMeta)
		if err != nil {
			return err
		}

		// Print queue
		fmt.Println(queue)
		return nil
	})
}

func (cmd QueueDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteQueue(ctx, cmd.Queue)
	})
}

func (cmd QueueUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Swap queue names
		cmd.Name, cmd.QueueMeta.Queue = cmd.QueueMeta.Queue, cmd.Name

		// Update the queue
		queue, err := provider.UpdateQueue(ctx, cmd.Name, cmd.QueueMeta)
		if err != nil {
			return err
		}

		// Print queue
		fmt.Println(queue)
		return nil
	})
}
