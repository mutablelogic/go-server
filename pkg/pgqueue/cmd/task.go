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

type TaskCommands struct {
	CreateTask  TaskCreateCommand  `cmd:"" group:"QUEUE" help:"Create a new task"`
	RetainTask  TaskRetainCommand  `cmd:"retain" group:"QUEUE" help:"Retain a task"`
	ReleaseTask TaskReleaseCommand `cmd:"release" group:"QUEUE" help:"Release a task"`
}

type TaskCreateCommand struct {
	Queue   string         `arg:"" help:"Queue name"`
	Payload *string        `help:"JSON Task payload"`
	Delay   *time.Duration `help:"Delay before task is executed"`
}

type TaskRetainCommand struct {
	Worker *string `help:"Worker name"`
}

type TaskReleaseCommand struct {
	Task  uint64  `arg:"" help:"Task ID"`
	Error *string `help:"Error message"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd TaskCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		meta := schema.TaskMeta{}
		if cmd.Payload != nil {
			if err := json.Unmarshal([]byte(types.PtrString(cmd.Payload)), &meta.Payload); err != nil {
				return fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}
		if cmd.Delay != nil {
			meta.DelayedAt = types.TimePtr(time.Now().Add(types.PtrDuration(cmd.Delay)))
		}

		// Create task
		task, err := provider.CreateTask(ctx, cmd.Queue, meta)
		if err != nil {
			return err
		}

		// Print task
		fmt.Println(task)
		return nil
	})
}

func (cmd TaskRetainCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Retain task
		task, err := provider.RetainTask(ctx, client.WithWorker(types.PtrString(cmd.Worker)))
		if err != nil {
			return err
		}

		// Print task
		fmt.Println(task)
		return nil
	})
}

func (cmd TaskReleaseCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		var err error
		if cmd.Error != nil {
			err = fmt.Errorf("task failed: %s", *cmd.Error)
		}

		// Release task
		task, err := provider.ReleaseTask(ctx, cmd.Task, err)
		if err != nil {
			return err
		}

		// Print task
		fmt.Println(task)
		return nil
	})
}
