package pgqueue

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// CreateTask creates a new task, and returns it.
func (client *Client) CreateTask(ctx context.Context, queue string, meta schema.TaskMeta) (*schema.Task, error) {
	var taskId schema.TaskId
	var task schema.TaskWithStatus
	if err := client.conn.With("id", queue).Insert(ctx, &taskId, meta); err != nil {
		return nil, err
	} else if err := client.conn.Get(ctx, &task, taskId); err != nil {
		return nil, err
	}
	return &task.Task, nil
}
