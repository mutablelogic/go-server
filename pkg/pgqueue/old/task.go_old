package pgqueue

import (
	"context"
	"errors"
	"sync"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
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

// GetTask returns a task based on identifier, and optionally sets the task status.
func (client *Client) GetTask(ctx context.Context, task uint64, status *string) (*schema.Task, error) {
	var taskObj schema.TaskWithStatus
	if err := client.conn.Get(ctx, &taskObj, schema.TaskId(task)); err != nil {
		return nil, err
	} else if status != nil {
		*status = taskObj.Status
	}
	return &taskObj.Task, nil
}

// NextTask retains a task, and returns it. Returns nil if there is no task to retain
func (client *Client) NextTask(ctx context.Context) (*schema.Task, error) {
	var taskId schema.TaskId
	var task schema.TaskWithStatus
	if err := client.conn.Get(ctx, &taskId, schema.TaskRetain{
		Worker: client.worker,
	}); err != nil {
		return nil, err
	} else if taskId == 0 {
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
		Id:     task,
		Fail:   false,
		Result: result,
	}); err != nil {
		return nil, err
	} else if taskId == 0 {
		// No task found
		return nil, pg.ErrNotFound
	}
	if err := client.conn.Get(ctx, &taskObj, taskId); err != nil {
		return nil, err
	}
	return &taskObj.Task, nil
}

// FailTask fails a task, either for retry or permanent failure, and returns the task and status.
func (client *Client) FailTask(ctx context.Context, task uint64, result any, status *string) (*schema.Task, error) {
	var taskId schema.TaskId
	var taskObj schema.TaskWithStatus
	if err := client.conn.Get(ctx, &taskId, schema.TaskRelease{
		Id:     task,
		Fail:   true,
		Result: result,
	}); err != nil {
		return nil, err
	} else if taskId == 0 {
		// No task found
		return nil, pg.ErrNotFound
	}
	if err := client.conn.Get(ctx, &taskObj, taskId); err != nil {
		return nil, err
	} else if status != nil {
		*status = taskObj.Status
	}
	return &taskObj.Task, nil
}

// RunTaskLoop runs a loop to process matured tasks, until the context is cancelled.
// It does not retain or release tasks, but simply returns them to the caller.
func (client *Client) RunTaskLoop(ctx context.Context, taskch chan<- *schema.Task, errch chan<- error) error {
	var wg sync.WaitGroup

	max_delta := schema.TaskPeriod
	min_delta := schema.TaskPeriod / 10
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()

	// Make a channel for notifications
	notifych := make(chan *pg.Notification)
	defer close(notifych)

	// Listen for notifications
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				notification, err := client.listener.WaitForNotification(ctx)
				if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					errch <- err
				} else {
					notifych <- notification
				}
			}
		}
	}(ctx)

	nextTaskFn := func() error {
		// Check for matured tasks
		task, err := client.NextTask(ctx)
		if err != nil {
			return err
		}
		if task != nil {
			// Emit task to channel
			taskch <- task
			// Retain task, reset timer to minimum period
			timer.Reset(min_delta)
		} else {
			// No task to retain, reset timer to maximum period
			timer.Reset(max_delta)
		}
		return nil
	}

	// Loop until context is cancelled
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case <-timer.C:
			if err := nextTaskFn(); err != nil {
				errch <- err
			}
		case <-notifych:
			if err := nextTaskFn(); err != nil {
				errch <- err
			}
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Return success
	return nil
}
