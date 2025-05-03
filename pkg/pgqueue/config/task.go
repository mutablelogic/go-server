package config

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	// Packages
	marshaler "github.com/djthorpe/go-marshaler"
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
	decoder   *marshaler.Decoder
}

var _ server.Task = (*task)(nil)
var _ server.PGQueue = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTask(manager *pgqueue.Manager, threads uint) (server.Task, error) {
	self := new(task)
	self.manager = manager
	self.taskpool = pgqueue.NewTaskPool(threads)
	self.callbacks = make(map[string]server.PGCallback, 100)
	self.decoder = marshaler.NewDecoder("json",
		convertPtr,
		convertPGTime,
		convertFloatToIntUint,
		marshaler.ConvertTime,
		marshaler.ConvertDuration,
		marshaler.ConvertIntUint,
	)
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
	ctx = ref.WithPath(ref.WithProvider(ctx, ref.Provider(parent)), ref.Path(parent)...)

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
			var namespace_queue []string
			if err := task.UnmarshalPayload(&namespace_queue, evt.Payload); err != nil {
				ref.Log(ctx).With("ticker", evt).Print(parent, err)
			} else if len(namespace_queue) == 2 && namespace_queue[0] == task.manager.Namespace() {
				n := 0
				for {
					tasks, err := task.manager.CleanQueue(ctx, namespace_queue[1])
					if err != nil {
						ref.Log(ctx).With("ticker", evt).Print(parent, "clean queue:", err)
						break
					}
					if len(tasks) == 0 {
						break
					}
					n += len(tasks)
				}
				if n > 0 {
					ref.Log(ctx).With("ticker", evt).Debug(parent, "removed ", n, " tasks from queue")
				}
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

// Conn returns the underlying connection pool object.
func (t *task) Conn() pg.PoolConn {
	return t.manager.Conn()
}

// Namespace returns the namespace of the queue.
func (t *task) Namespace() string {
	return t.manager.Namespace()
}

// RegisterTicker registers a periodic task (ticker) with a callback function.
// It returns the metadata of the registered ticker.
func (t *task) RegisterTicker(ctx context.Context, meta schema.TickerMeta, fn server.PGCallback) (*schema.Ticker, error) {
	ref.Log(ctx).Debug(ctx, "Register ticker: ", meta.Ticker, " in namespace ", t.manager.Namespace())
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
	ref.Log(ctx).Debug(ctx, "Register queue: ", meta.Queue, " in namespace ", t.manager.Namespace())
	queue, err := t.manager.RegisterQueue(ctx, meta)
	if err != nil {
		return nil, err
	}

	// Register a queue cleanup timer
	if ticker, err := t.manager.RegisterTickerNs(ctx, schema.CleanupNamespace, schema.TickerMeta{
		Ticker:   joinName(queue.Namespace, queue.Queue),
		Payload:  []string{queue.Namespace, queue.Queue},
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

// Convert a payload into a struct
func (t *task) UnmarshalPayload(dest any, payload any) error {
	return t.decoder.Decode(payload, dest)
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
		child := ref.WithPath(ref.WithProvider(context.TODO(), ref.Provider(ctx)), ref.Path(ctx)...)
		if _, err_ := t.manager.ReleaseTask(child, task.Id, err == nil, err, &status); err_ != nil {
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

const namespaceSeparator = "_"

func joinName(parts ...string) string {
	return strings.Join(parts, namespaceSeparator)
}

// //////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

var (
	nilValue = reflect.ValueOf(nil)
	timeType = reflect.TypeOf(time.Time{})
)

// convertTime returns time in postgres format
func convertPGTime(src reflect.Value, dest reflect.Type) (reflect.Value, error) {
	// Pass value through
	if src.Type() == dest {
		return src, nil
	}

	if dest == timeType {
		// Convert time 2025-05-03T17:29:32.329803 => time.Time
		if t, err := time.Parse("2006-01-02T15:04:05.999999999", src.String()); err == nil {
			return reflect.ValueOf(t), nil
		}
	}

	// Skip
	return nilValue, nil
}

// convertPtr returns value if pointer
func convertPtr(src reflect.Value, dest reflect.Type) (reflect.Value, error) {
	// Pass value through
	if src.Type() == dest {
		return src, nil
	}

	// Convert src to elem
	if dest.Kind() == reflect.Ptr {
		if dest.Elem() == src.Type() {
			if src.CanAddr() {
				return src.Addr(), nil
			} else {
				src_ := reflect.New(dest.Elem())
				src_.Elem().Set(src)
				return src_, nil
			}
		}
	}

	// Skip
	return nilValue, nil
}

// convert float types to int or uint
func convertFloatToIntUint(src reflect.Value, dest reflect.Type) (reflect.Value, error) {
	// Pass value through
	if src.Type() == dest {
		return src, nil
	}
	switch dest.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src.Kind() == reflect.Float64 || src.Kind() == reflect.Float32 {
			return reflect.ValueOf(int64(src.Float())), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src.Kind() == reflect.Float64 || src.Kind() == reflect.Float32 {
			return reflect.ValueOf(uint64(src.Float())), nil
		}
	}
	// Skip
	return nilValue, nil
}
