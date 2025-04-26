package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	sync.WaitGroup
	client  *pgqueue.Client
	tickers map[string]exec
	queues  map[string]exec
}

type exec struct {
	deadline time.Duration
	fn       server.PGCallback
}

var _ server.Task = (*task)(nil)
var _ server.PGQueue = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func taskWith(queue *pgqueue.Client) *task {
	return &task{
		client:  queue,
		tickers: make(map[string]exec, 10),
		queues:  make(map[string]exec, 10),
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *task) Run(ctx context.Context) error {
	var result error

	// Defer closing the listener
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result = errors.Join(result, t.client.Close(ctx))
	}()

	// Create a ticker channel and defer close it
	tickerch := make(chan *schema.Ticker)
	defer close(tickerch)

	// Create an error channel and defer close it
	errch := make(chan error)
	defer close(errch)

	// Emit ticker events until the context is cancelled
	t.Add(1)
	go func(ctx context.Context) {
		defer t.Done()
		result = errors.Join(result, t.client.RunTickerLoop(ctx, tickerch))
	}(ctx)

	// Create a task channel and defer close it
	taskch := make(chan *schema.Task)
	defer close(taskch)

	// Emit task events until the context is cancelled
	t.Add(1)
	go func(ctx context.Context) {
		defer t.Done()
		result = errors.Join(result, t.client.RunTaskLoop(ctx, taskch, errch))
	}(ctx)

	// Continue until context is done
	parent, parentCancel := context.WithCancel(context.Background())
	go func(ctx context.Context, parentCancel context.CancelFunc) {
		// Wait for the context to be done
		<-ctx.Done()
		// Wait for the goroutines to finish
		t.Wait()
		// Cancel the runLoop
		parentCancel()
	}(ctx, parentCancel)

	// Runloop until the parent context is done
FOR_LOOP:
	for {
		select {
		case <-parent.Done():
			// Break out of the loop
			break FOR_LOOP
		case ticker := <-tickerch:
			t.execTicker(ctx, ticker, errch)
		case task := <-taskch:
			t.execTask(ctx, task, errch)
		case err := <-errch:
			// TODO: Handle errors, for now, just log them
			log.Println("ERROR:", err)
		}
	}

	// Return any errors
	return result
}

// Return the worker name
func (t *task) Worker() string {
	return t.client.Worker()
}

// Register a ticker with a callback, and return the registered ticker
func (t *task) RegisterTicker(ctx context.Context, meta schema.TickerMeta, fn server.PGCallback) (*schema.Ticker, error) {
	ticker, err := t.client.RegisterTicker(ctx, meta)
	if err != nil {
		return nil, err
	}

	// Add the ticker to the map
	t.tickers[ticker.Ticker] = exec{
		deadline: types.PtrDuration(ticker.Interval),
		fn:       fn,
	}

	// Return success
	return ticker, nil
}

// Delete an existing ticker
func (t *task) DeleteTicker(ctx context.Context, name string) error {
	ticker, err := t.client.DeleteTicker(ctx, name)
	if err != nil {
		return err
	}

	// Remove the ticker from the map
	delete(t.tickers, ticker.Ticker)

	// Return success
	return nil
}

// Register a queue with a callback, and return the registered queue
func (t *task) RegisterQueue(ctx context.Context, meta schema.Queue, fn server.PGCallback) (*schema.Queue, error) {
	queue, err := t.client.RegisterQueue(ctx, meta)
	if err != nil {
		return nil, err
	}

	// Add the queue to the map
	t.queues[queue.Queue] = exec{
		deadline: types.PtrDuration(queue.TTL),
		fn:       fn,
	}

	// Return success
	return queue, nil
}

// Push a task onto a queue, and return the task
func (t *task) CreateTask(ctx context.Context, queue string, payload any, delay time.Duration) (*schema.Task, error) {
	var delayedAt *time.Time
	if delay > 0 {
		delayedAt = types.TimePtr(time.Now().Add(delay))
	}
	return t.client.CreateTask(ctx, queue, schema.TaskMeta{
		Payload:   payload,
		DelayedAt: delayedAt,
	})
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func contextWithDeadline(ctx context.Context, deadline time.Duration) (context.Context, context.CancelFunc) {
	if deadline > 0 {
		return context.WithTimeout(ctx, deadline)
	}
	return ctx, func() {}
}

// Execute a callback function for a ticker or queue
func (t *task) exec(ctx context.Context, fn exec, in any) (result error) {
	// Create a context with a deadline
	deadline, cancel := contextWithDeadline(ctx, fn.deadline)
	defer cancel()

	// Catch panics
	defer func() {
		if r := recover(); r != nil {
			result = errors.Join(result, fmt.Errorf("PANIC: %v", r))
		}
	}()

	// Run the ticker or task handler
	err := fn.fn(deadline, in)
	if err != nil {
		result = errors.Join(result, err)
	}

	// Concatenate any errors from the deadline
	if deadline.Err() != nil && !errors.Is(err, deadline.Err()) {
		result = errors.Join(result, deadline.Err())
	}

	// Return any errors
	return result
}

// Execute a ticker callback
func (t *task) execTicker(ctx context.Context, ticker *schema.Ticker, errch chan<- error) {
	fn, exists := t.tickers[ticker.Ticker]
	if !exists {
		return
	}

	// Execute the callback function in a goroutine
	t.Add(1)
	go func(ctx context.Context) {
		defer t.Done()
		if err := t.exec(ctx, fn, nil); err != nil {
			errch <- fmt.Errorf("TICKER %q: %w", ticker.Ticker, err)
		}
	}(contextWithTicker(ctx, ticker))
}

// Execute a task callback
func (t *task) execTask(ctx context.Context, task *schema.Task, errch chan<- error) {
	fn, exists := t.queues[task.Queue]
	if !exists {
		return
	}

	// Execute the callback function in a goroutine - and then release or fail the task
	t.Add(1)
	go func(ctx context.Context) {
		defer t.Done()
		if err := t.exec(ctx, fn, task.Payload); err != nil {
			// TODO: Fail task
			errch <- fmt.Errorf("TASK %q_%d: %w", task.Queue, task.Id, err)
		} else {
			// TODO: Release task
		}
	}(contextWithTask(ctx, task))
}
