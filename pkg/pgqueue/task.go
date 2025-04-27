package pgqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	ref "github.com/mutablelogic/go-server/pkg/ref"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RunTicker(ctx context.Context, ticker *schema.Ticker, fn server.PGCallback) error {
	var deadline time.Duration
	if ticker.Interval != nil {
		deadline = *ticker.Interval
	}
	return runTask(ref.WithTicker(ctx, ticker), deadline, fn, nil)
}

func RunTask(ctx context.Context, task *schema.Task, fn server.PGCallback, payload any) error {
	var deadline time.Duration
	if task.DiesAt != nil {
		deadline = time.Until(*task.DiesAt)
	}
	return runTask(ref.WithTask(ctx, task), deadline, fn, payload)
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

	// Concatenate any errors from the deadline
	if ctx.Err() != nil {
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
