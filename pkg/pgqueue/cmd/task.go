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
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TaskCommands struct {
	CreateTask TaskCreateCommand `cmd:"" group:"QUEUE" help:"Create a new task"`
}

type TaskCreateCommand struct {
	Queue   string         `arg:"" help:"Queue name"`
	Payload *string        `help:"JSON Task payload"`
	Delay   *time.Duration `help:"Delay before task is executed"`
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
