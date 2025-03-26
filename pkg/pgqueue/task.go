package pgqueue

import (
	"context"
	"fmt"

	// Packages

	"github.com/djthorpe/go-pg"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	"github.com/mutablelogic/go-server/pkg/types"
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

// RetainTask retains a task from a queue, and returns it. Returns nil if there is no
// task to retain
func (client *Client) RetainTask(ctx context.Context, queue string) (*schema.Task, error) {
	var taskId schema.TaskId
	var task schema.TaskWithStatus
	if err := client.conn.Get(ctx, &taskId, schema.TaskRetain{
		Queue:  queue,
		Worker: client.worker,
	}); err != nil {
		return nil, err
	} else if taskId.Id == nil {
		// No task to retain
		return nil, nil
	} else if err := client.conn.Get(ctx, &task, taskId); err != nil {
		return nil, err
	}
	return &task.Task, nil
}

// ReleaseTask releases a task from a queue, and returns it.
func (client *Client) ReleaseTask(ctx context.Context, task uint64, result any) (*schema.Task, error) {
	var taskId schema.TaskId
	var taskObj schema.TaskWithStatus
	if err := client.conn.Get(ctx, &taskId, schema.TaskRelease{
		TaskId: schema.TaskId{Id: types.Uint64Ptr(task)},
		Fail:   false,
		Result: result,
	}); err != nil {
		return nil, err
	} else if taskId.Id == nil {
		// No task found
		return nil, pg.ErrNotFound
	}
	fmt.Println("taskId", taskId)
	if err := client.conn.Get(ctx, &taskObj, taskId); err != nil {
		return nil, err
	}
	return &taskObj.Task, nil
}
