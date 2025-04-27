package pgqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type TaskPool struct {
	wg      sync.WaitGroup
	workers chan task
}

type task struct {
	ctx      context.Context
	deadline time.Duration
	fn       server.PGCallback
	result   func(error)
	payload  any
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTaskPool(workers uint) *TaskPool {
	self := new(TaskPool)
	self.workers = make(chan task, workers)

	// Create the workers
	for i := uint(0); i < workers; i++ {
		self.wg.Add(1)
		go func() {
			defer self.wg.Done()
			for task := range self.workers {
				task.result(runTask(task.ctx, task.deadline, task.fn, task.payload))
			}
		}()
	}

	// Return the task pool
	return self
}

func (pool *TaskPool) Close() {
	// Close the workers channel
	close(pool.workers)

	// Wait for all tasks to finish
	pool.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Queue a ticker for running. Will block whilst all the workers are busy
func (pool *TaskPool) RunTicker(ctx context.Context, t *schema.Ticker, fn server.PGCallback, result func(error)) {
	var deadline time.Duration
	if t.Interval != nil {
		deadline = types.PtrDuration(t.Interval)
	}
	if result == nil {
		result = func(err error) {}
	}
	pool.workers <- task{
		ctx:      ref.WithTicker(ctx, t),
		deadline: deadline,
		fn:       fn,
		result:   result,
		payload:  nil,
	}
}

// Queue a task for running. Will block whilst all the workers are busy
func (pool *TaskPool) RunTask(ctx context.Context, t *schema.Task, fn server.PGCallback, payload any, result func(error)) {
	var deadline time.Duration
	if t.DiesAt != nil {
		deadline = time.Until(*t.DiesAt)
	}
	if result == nil {
		result = func(err error) {}
	}
	pool.workers <- task{
		ctx:      ref.WithTask(ctx, t),
		deadline: deadline,
		fn:       fn,
		result:   result,
		payload:  payload,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// RunTask runs a task with the given context and payload. The task will be
// executed with a deadline. If the task does not complete within the deadline,
// it will be cancelled and an error will be returned.
func runTask(parent context.Context, deadline time.Duration, fn server.PGCallback, payload any) (errs error) {
	ctx, cancel := contextWithDeadline(parent, deadline)
	defer cancel()

	// Catch panics
	defer func() {
		if r := recover(); r != nil {
			errs = errors.Join(errs, fmt.Errorf("PANIC: %v", r))
		}
	}()

	// Run the task function
	if err := fn(ctx, payload); err != nil {
		errs = errors.Join(errs, err)
	}

	// Concatenate any errors from the context if not already present
	if ctx.Err() != nil && !errors.Is(errs, ctx.Err()) {
		errs = errors.Join(errs, ctx.Err())
	}

	// Return errs
	return errs
}

func contextWithDeadline(ctx context.Context, deadline time.Duration) (context.Context, context.CancelFunc) {
	if deadline > 0 {
		return context.WithTimeout(ctx, deadline)
	}
	return ctx, func() {}
}
