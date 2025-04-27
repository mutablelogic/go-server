package config

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	manager *pgqueue.Manager
	wg      sync.WaitGroup
	workers uint
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *task) Run(parent context.Context) error {
	var errs error

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	ctx = ref.WithPath(ref.WithLog(ctx, ref.Log(parent)), ref.Path(parent)...)

	// Ticker loop
	tickerch := make(chan *schema.Ticker)
	defer close(tickerch)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunTickerLoop(ctx, task.manager.Namespace(), tickerch); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

	// Ticker loop for cleanup of tasks
	cleanupch := make(chan *schema.Ticker)
	defer close(cleanupch)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunTickerLoop(ctx, schema.CleanupNamespace, cleanupch); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

	// Notification loop
	notifych := make(chan *pg.Notification)
	defer close(notifych)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunNotificationLoop(ctx, notifych); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

	// Task loop
	taskch := make(chan *schema.Task)
	defer close(taskch)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunTaskLoop(ctx, taskch); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

	// Task worker pool
	taskpool := pgqueue.NewTaskPool(task.workers)
	ref.Log(ctx).Debug(parent, "Created task pool with ", task.workers, " threads")

FOR_LOOP:
	for {
		select {
		case <-parent.Done():
			cancel()
			break FOR_LOOP
		case <-notifych:
			// Handle notification - try and retain a tasks
			for {
				evt, err := task.manager.NextTask(ctx)
				if err != nil {
					ref.Log(ctx).Print(parent, err)
					break
				}
				if evt == nil {
					break
				}
				task.tryTask(ctx, taskpool, evt)
			}
		case evt := <-tickerch:
			task.tryTicker(ctx, taskpool, evt)
		case evt := <-taskch:
			task.tryTask(ctx, taskpool, evt)
		case evt := <-cleanupch:
			ref.Log(ctx).Print(parent, "CLEANUP TICKER ", evt)
		}
	}

	// Wait for all goroutines to finish
	task.wg.Wait()

	// Return any errors
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterTicker registers a periodic task (ticker) with a callback function.
// It returns the metadata of the registered ticker.
func (t *task) RegisterTicker(ctx context.Context, meta schema.TickerMeta, fn server.PGCallback) (*schema.Ticker, error) {
	ticker, err := t.manager.RegisterTicker(ctx, meta)
	if err != nil {
		return nil, err
	}
	// TODO: Register the ticker callback
	return ticker, nil
}

// RegisterQueue registers a task queue with a callback function.
// It returns the metadata of the registered queue.
func (t *task) RegisterQueue(ctx context.Context, meta schema.QueueMeta, fn server.PGCallback) (*schema.Queue, error) {
	queue, err := t.manager.RegisterQueue(ctx, meta)
	if err != nil {
		return nil, err
	}
	// Register a queue cleanup timer
	if _, err := t.manager.RegisterTickerNs(ctx, schema.CleanupNamespace, schema.TickerMeta{
		Ticker:   queue.Queue,
		Interval: queue.TTL,
	}); err != nil {
		_, err_ := t.manager.DeleteQueue(ctx, meta.Queue)
		return nil, errors.Join(err, err_)
	}
	return queue, nil
}

// CreateTask adds a new task to a specified queue with a payload and optional delay.
// It returns the metadata of the created task.
func (t *task) CreateTask(ctx context.Context, queue string, payload any, delay time.Duration) (*schema.Task, error) {
	meta := schema.TaskMeta{
		Payload: payload,
	}
	if delay > 0 {
		meta.DelayedAt = types.TimePtr(time.Now().Add(delay))
	}
	return t.manager.CreateTask(ctx, queue, meta)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t *task) tryTicker(ctx context.Context, taskpool *pgqueue.TaskPool, ticker *schema.Ticker) {
	now := time.Now()
	taskpool.RunTicker(ctx, ticker, tickerFunc, func(err error) {
		delta := time.Since(now).Truncate(time.Millisecond)
		switch {
		case err == nil:
			ref.Log(ctx).With("ticker", ticker, "delta_ms", delta.Milliseconds()).Print(ctx, "Ticker completed in ", delta)
		default:
			ref.Log(ctx).With("ticker", ticker, "delta_ms", delta.Milliseconds(), "error", err.Error()).Printf(ctx, "Failed after %v: %v", delta, err)
		}
	})
}

func (t *task) tryTask(ctx context.Context, taskpool *pgqueue.TaskPool, task *schema.Task) {
	now := time.Now()
	taskpool.RunTask(ctx, task, taskFunc, task.Payload, func(err error) {
		var status string
		delta := time.Since(now).Truncate(time.Millisecond)
		if _, err_ := t.manager.ReleaseTask(context.TODO(), task.Id, err == nil, err, &status); err_ != nil {
			err = errors.Join(err, err_)
		}
		switch {
		case err == nil:
			ref.Log(ctx).With("task", task, "delta_ms", delta.Milliseconds(), "status", "success").Print(ctx, "Task completed in ", delta)
		default:
			ref.Log(ctx).With("task", task, "delta_ms", delta.Milliseconds(), "status", status, "error", err.Error()).Printf(ctx, "Failed (with %s) after %v: %v", status, delta, err)
		}
	})
}

////////////////////////////////////////////////////////////////////////////////
// TEST CALLBACKS

func taskFunc(ctx context.Context, payload any) error {
	var err error
	if rand.Intn(2) == 1 {
		err = errors.New("random error")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * time.Duration(rand.Intn(10))):
		return err
	}
}

func tickerFunc(ctx context.Context, payload any) error {
	var err error
	if rand.Intn(2) == 1 {
		err = errors.New("random error")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * time.Duration(rand.Intn(20))):
		return err
	}
}
