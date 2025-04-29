package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ref "github.com/mutablelogic/go-server/pkg/ref"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	providerLabel = "provider"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run a task until context is cancelled and return any errors
func (provider *provider) Run(parent context.Context) error {
	var result error

	// Append the provider to the context
	parent = ref.WithPath(ref.WithProvider(parent, provider), providerLabel)

	// Create a child context which will allow us to cancel all the tasks
	// prematurely if any of them fail
	ctx, prematureCancel := context.WithCancel(parent)
	defer prematureCancel()

	// Create all the tasks
	if err := provider.constructor(parent); err != nil {
		return err
	}

	// Run the tasks in parallel
	for _, label := range provider.order {
		task, exists := provider.task[label]
		if !exists {
			return httpresponse.ErrNotFound.Withf("Task %q not found", label)
		}

		// Create a context for each task
		ctx, cancel := context.WithCancel(context.Background())
		task.Context = ref.WithPath(ref.WithProvider(ctx, provider), label)
		task.CancelFunc = cancel
		task.WaitGroup.Add(1)
		go func(task *state) {
			defer task.WaitGroup.Done()
			defer task.CancelFunc()
			provider.Print(parent, "Starting task ", label)
			if err := task.Run(task.Context); err != nil {
				result = errors.Join(result, fmt.Errorf("%q: %w", label, err))
				prematureCancel()
			}
		}(task)
	}

	provider.Print(parent, "Press CTRL+C to cancel all tasks")

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Cancel all the tasks in reverse order, waiting for each to complete before cancelling the next
	sort.Sort(sort.Reverse(sort.StringSlice(provider.order)))
	for _, label := range provider.order {
		provider.Print(ctx, "Stopping task ", label)
		provider.task[label].CancelFunc()
		provider.task[label].WaitGroup.Wait()
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Make all tasks
func (provider *provider) constructor(ctx context.Context) error {
	for _, label := range provider.porder {
		if task := provider.Task(ctx, label); task == nil {
			return httpresponse.ErrConflict.Withf("Failed to create task %q", label)
		}
	}
	return nil
}
