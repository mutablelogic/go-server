package config

import (
	"context"
	"errors"
	"sync"

	// Packages
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	ref "github.com/mutablelogic/go-server/pkg/ref"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	manager *pgqueue.Manager
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *task) Run(parent context.Context) error {
	var wg sync.WaitGroup
	var errs error

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	ctx = ref.WithPath(ref.WithLog(ctx, ref.Log(parent)), ref.Label(parent))

	// Ticker loop
	tickerch := make(chan *schema.Ticker)
	defer close(tickerch)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := task.manager.RunTickerLoop(ctx, tickerch); err != nil {
			errs = errors.Join(errs, err)
		}

	}()

	// Task loop
	taskch := make(chan *schema.Task)
	defer close(taskch)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := task.manager.RunTaskLoop(ctx, taskch); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

FOR_LOOP:
	for {
		select {
		case <-parent.Done():
			cancel()
			break FOR_LOOP
		case evt := <-tickerch:
			// Handle ticker
			ref.Log(ctx).Debug(parent, "TICKER", evt)
		case evt := <-taskch:
			// Handle task
			ref.Log(ctx).Debug(parent, "TRY TASK", evt)
			// TODO: Fail task
			var status string
			task.manager.ReleaseTask(parent, evt.Id, false, "Task failed", &status)
			ref.Log(ctx).Debug(parent, "FAILED STATUS => ", status)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Return any errors
	return nil
}
