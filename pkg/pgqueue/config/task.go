package config

import (
	"context"
	"errors"
	"strings"
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
	wg        sync.WaitGroup
	manager   *pgqueue.Manager
	taskpool  *pgqueue.TaskPool
	callbacks map[string]server.PGCallback
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTask(manager *pgqueue.Manager, threads uint) (server.Task, error) {
	self := new(task)
	self.manager = manager
	self.taskpool = pgqueue.NewTaskPool(threads)
	self.callbacks = make(map[string]server.PGCallback, 100)
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *task) Run(parent context.Context) error {
	var errs error

	// Close the task pool when done
	defer task.taskpool.Close()

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
				task.tryTask(ctx, task.taskpool, evt)
			}
		case evt := <-tickerch:
			task.tryTicker(ctx, task.taskpool, evt)
		case evt := <-taskch:
			task.tryTask(ctx, task.taskpool, evt)
		case evt := <-cleanupch:
			if namespace_queue := splitName(evt.Ticker, 2); len(namespace_queue) == 2 {
				ref.Log(ctx).Print(parent, "CLEANUP TICKER ns=", namespace_queue[0], " queue=", namespace_queue[1])
			}
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
	ref.Log(ctx).Print(ctx, "Register ticker: ", meta.Ticker)
	ticker, err := t.manager.RegisterTicker(ctx, meta)
	if err != nil {
		return nil, err
	}

	// Register the ticker callback
	t.setTickerCallback(ticker, fn)

	// Return the ticker metadata
	return ticker, nil
}

// RegisterQueue registers a task queue with a callback function.
// It returns the metadata of the registered queue.
func (t *task) RegisterQueue(ctx context.Context, meta schema.QueueMeta, fn server.PGCallback) (*schema.Queue, error) {
	ref.Log(ctx).Print(ctx, "Register queue: ", meta.Queue)
	queue, err := t.manager.RegisterQueue(ctx, meta)
	if err != nil {
		return nil, err
	}

	// Register a queue cleanup timer
	if ticker, err := t.manager.RegisterTickerNs(ctx, schema.CleanupNamespace, schema.TickerMeta{
		Ticker:   joinName(queue.Namespace, queue.Queue),
		Interval: queue.TTL,
	}); err != nil {
		_, err_ := t.manager.DeleteQueue(ctx, meta.Queue)
		return nil, errors.Join(err, err_)
	} else {
		t.setTickerCallback(ticker, func(ctx context.Context, _ any) error {
			// Cleanup the queue
			if _, err := t.manager.CleanQueue(ctx, queue.Queue); err != nil {
				return err
			}
			return nil
		})
	}

	// Register the task callback
	t.setTaskCallback(queue, fn)

	// Return the queue metadata
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
	taskpool.RunTicker(ctx, ticker, t.getTickerCallback(ticker), func(err error) {
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
	taskpool.RunTask(ctx, task, t.getTaskCallback(task), func(err error) {
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

func (t *task) setTickerCallback(ticker *schema.Ticker, fn server.PGCallback) {
	key := joinName("ticker", ticker.Namespace, ticker.Ticker)
	t.callbacks[key] = fn
}

func (t *task) setTaskCallback(queue *schema.Queue, fn server.PGCallback) {
	key := joinName("task", queue.Namespace, queue.Queue)
	t.callbacks[key] = fn
}

func (t *task) getTickerCallback(ticker *schema.Ticker) server.PGCallback {
	key := joinName("ticker", ticker.Namespace, ticker.Ticker)
	return t.callbacks[key]
}

func (t *task) getTaskCallback(task *schema.Task) server.PGCallback {
	key := joinName("task", task.Namespace, task.Queue)
	return t.callbacks[key]
}

const namespaceSeparator = "/"

func joinName(parts ...string) string {
	return strings.Join(parts, namespaceSeparator)
}

func splitName(name string, n int) []string {
	return strings.SplitN(name, namespaceSeparator, n)
}
