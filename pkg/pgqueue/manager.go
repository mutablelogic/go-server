package pgqueue

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	// Packages

	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	conn      pg.PoolConn
	worker    string
	namespace string
	listener  pg.Listener
	topics    []string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*Manager, error) {
	self := new(Manager)
	opts, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}

	// Set worker name
	if worker, err := defaultWorker(); err != nil {
		return nil, err
	} else if opts.worker == "" {
		self.worker = worker
	} else {
		self.worker = opts.worker
	}
	self.namespace = opts.namespace

	// Set the connection
	self.conn = conn.With(
		"schema", schema.SchemaName,
		"ns", opts.namespace,
	).(pg.PoolConn)

	// Create the schema
	if exists, err := pg.SchemaExists(ctx, conn, schema.SchemaName); err != nil {
		return nil, err
	} else if !exists {
		if err := pg.SchemaCreate(ctx, conn, schema.SchemaName); err != nil {
			return nil, err
		}
	}

	// Bootstrap the schema
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		return schema.Bootstrap(ctx, conn)
	}); err != nil {
		return nil, err
	}

	// Set the listener and topics
	self.listener = self.conn.Listener()
	self.topics = []string{
		strings.Join([]string{opts.namespace, schema.TopicQueueInsert}, "_"),
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (manager *Manager) Namespace() string {
	return manager.namespace
}

func (manager *Manager) Worker() string {
	return manager.worker
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TICKER

// RegisterTicker creates a new ticker, or updates an existing ticker, and returns it.
func (manager *Manager) RegisterTicker(ctx context.Context, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get a ticker
		if err := conn.Get(ctx, &ticker, schema.TickerName(meta.Ticker)); err != nil && !errors.Is(err, pg.ErrNotFound) {
			return err
		} else if errors.Is(err, pg.ErrNotFound) {
			// If the ticker does not exist, then create it
			if err := conn.Insert(ctx, &ticker, meta); err != nil {
				return err
			}
		}

		// Finally, update the ticker
		return conn.Update(ctx, &ticker, schema.TickerName(meta.Ticker), meta)
	}); err != nil {
		return nil, httperr(err)
	}
	return &ticker, nil
}

// RegisterTicker creates a new ticker, or updates an existing ticker, and returns it.
func (manager *Manager) RegisterTickerNs(ctx context.Context, namespace string, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker

	// Check namespace is valid
	if !types.IsIdentifier(namespace) {
		return nil, httpresponse.ErrBadRequest.Withf("Invalid namespace %q", namespace)
	}

	// Register the ticker
	if err := manager.conn.With("ns", namespace).Tx(ctx, func(conn pg.Conn) error {
		// Get a ticker
		if err := conn.Get(ctx, &ticker, schema.TickerName(meta.Ticker)); err != nil && !errors.Is(err, pg.ErrNotFound) {
			return err
		} else if errors.Is(err, pg.ErrNotFound) {
			// If the ticker does not exist, then create it
			if err := conn.Insert(ctx, &ticker, meta); err != nil {
				return err
			}
		}

		// Finally, update the ticker
		return conn.Update(ctx, &ticker, schema.TickerName(meta.Ticker), meta)
	}); err != nil {
		return nil, httperr(err)
	}
	return &ticker, nil
}

// UpdateTicker updates an existing ticker, and returns it.
func (manager *Manager) UpdateTicker(ctx context.Context, name string, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := manager.conn.Update(ctx, &ticker, schema.TickerName(name), meta); err != nil {
		return nil, httperr(err)
	}
	return &ticker, nil
}

// GetTicker returns a ticker by name
func (manager *Manager) GetTicker(ctx context.Context, name string) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := manager.conn.Get(ctx, &ticker, schema.TickerName(name)); err != nil {
		return nil, httperr(err)
	}
	return &ticker, nil
}

// DeleteTicker deletes an existing ticker, and returns the deleted ticker.
func (manager *Manager) DeleteTicker(ctx context.Context, name string) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		return conn.Delete(ctx, &ticker, schema.TickerName(name))
	}); err != nil {
		return nil, httperr(err)
	}
	return &ticker, nil
}

// ListTickers returns all tickers in a namespace as a list
func (manager *Manager) ListTickers(ctx context.Context, req schema.TickerListRequest) (*schema.TickerList, error) {
	var list schema.TickerList
	if err := manager.conn.List(ctx, &list, req); err != nil {
		return nil, err
	}
	return &list, nil
}

// NextTickerNs returns the next matured ticker in a namespace, or nil
func (manager *Manager) NextTickerNs(ctx context.Context, namespace string) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := manager.conn.With("ns", namespace).Get(ctx, &ticker, schema.TickerNext{}); errors.Is(err, pg.ErrNotFound) {
		// No matured ticker
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Return matured ticker
	return &ticker, nil
}

// NextTicker returns the next matured ticker, or nil
func (manager *Manager) NextTicker(ctx context.Context) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := manager.conn.Get(ctx, &ticker, schema.TickerNext{}); errors.Is(err, pg.ErrNotFound) {
		// No matured ticker
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Return matured ticker
	return &ticker, nil
}

// RunTickerLoop runs a loop to process matured tickers, until the context is cancelled,
// or an error occurs.
func (manager *Manager) RunTickerLoop(ctx context.Context, namespace string, ch chan<- *schema.Ticker) error {
	delta := schema.TickerPeriod
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	prev := time.Now()

	// Loop until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			// Check for matured tickers
			ticker, err := manager.NextTickerNs(ctx, namespace)
			if err != nil {
				return err
			}

			if ticker != nil {
				ch <- ticker

				// Reset timer to minimum period
				if dur := types.PtrDuration(ticker.Interval); dur >= time.Second && dur < delta {
					delta = dur
				}
				// Adjust based on time since last ticker
				if since := time.Since(prev); since > delta {
					delta += (since - delta) / 2
				} else if since < delta {
					delta -= (delta - since) / 2
				}

				// Reset the timer
				prev = time.Now()
			}

			// Next loop
			timer.Reset(delta)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUEUE

// RegisterQueue creates a new queue, or updates an existing queue, and returns it.
func (manager *Manager) RegisterQueue(ctx context.Context, meta schema.QueueMeta) (*schema.Queue, error) {
	var queue schema.Queue
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get a queue
		if err := conn.Get(ctx, &queue, schema.QueueName(meta.Queue)); err != nil && !errors.Is(err, pg.ErrNotFound) {
			return err
		} else if errors.Is(err, pg.ErrNotFound) {
			// If the queue does not exist, then create it
			if err := conn.Insert(ctx, &queue, meta); err != nil {
				return err
			}
		}

		// Update the queue
		return conn.Update(ctx, &queue, schema.QueueName(meta.Queue), meta)
	}); err != nil {
		return nil, httperr(err)
	}

	return &queue, nil
}

// ListQueues returns all queues in a namespace as a list
func (manager *Manager) ListQueues(ctx context.Context, req schema.QueueListRequest) (*schema.QueueList, error) {
	var list schema.QueueList
	if err := manager.conn.List(ctx, &list, req); err != nil {
		return nil, httperr(err)
	}
	return &list, nil
}

// GetQueue returns a queue by name
func (manager *Manager) GetQueue(ctx context.Context, name string) (*schema.Queue, error) {
	var queue schema.Queue
	if err := manager.conn.Get(ctx, &queue, schema.QueueName(name)); err != nil {
		return nil, httperr(err)
	}
	return &queue, nil
}

// DeleteQueue deletes an existing queue, and returns it
func (manager *Manager) DeleteQueue(ctx context.Context, name string) (*schema.Queue, error) {
	var queue schema.Queue
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		return conn.Delete(ctx, &queue, schema.QueueName(name))
	}); err != nil {
		return nil, httperr(err)
	}
	return &queue, nil
}

// UpdateQueue updates an existing queue, and returns it.
func (manager *Manager) UpdateQueue(ctx context.Context, name string, meta schema.QueueMeta) (*schema.Queue, error) {
	var queue schema.Queue
	if err := manager.conn.Update(ctx, &queue, schema.QueueName(name), meta); err != nil {
		return nil, httperr(err)
	}
	return &queue, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TASK

// CreateTask creates a new task, and returns it.
func (manager *Manager) CreateTask(ctx context.Context, queue string, meta schema.TaskMeta) (*schema.Task, error) {
	var taskId schema.TaskId
	var task schema.TaskWithStatus

	// Insert the task, and return it
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := conn.With("id", queue).Insert(ctx, &taskId, meta); err != nil {
			return err
		}
		return conn.Get(ctx, &task, taskId)
	}); err != nil {
		return nil, httperr(err)
	}

	// Return the task
	return &task.Task, nil
}

// NextTask retains a task, and returns it. Returns nil if there is no task to retain
func (manager *Manager) NextTask(ctx context.Context, opt ...Opt) (*schema.Task, error) {
	var taskId schema.TaskId
	var task schema.TaskWithStatus

	// Get worker name
	opts, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}
	if opts.worker == "" {
		opts.worker = manager.worker
	}

	// Insert the task, and return it
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := conn.Get(ctx, &taskId, schema.TaskRetain{
			Worker: opts.worker,
		}); err != nil {
			return err
		}

		// No task to retain
		if taskId == 0 {
			return nil
		}

		// Get the task
		return conn.Get(ctx, &task, taskId)
	}); err != nil {
		return nil, httperr(err)
	}

	// Return task
	if taskId == 0 {
		return nil, nil
	} else {
		return &task.Task, nil
	}
}

// ReleaseTask releases a task from a queue, and returns it. Can optionally set the status
func (manager *Manager) ReleaseTask(ctx context.Context, task uint64, success bool, result any, status *string) (*schema.Task, error) {
	var taskId schema.TaskId
	var taskObj schema.TaskWithStatus

	// Release the task, and return it
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := conn.Get(ctx, &taskId, schema.TaskRelease{Id: task, Fail: !success, Result: result}); err != nil {
			return err
		}

		// No task found
		if taskId == 0 {
			return pg.ErrNotFound
		}

		// Get the task
		return conn.Get(ctx, &taskObj, taskId)
	}); err != nil {
		return nil, httperr(err)
	}

	// Optionally set the status
	if status != nil {
		*status = taskObj.Status
	}

	// Return task
	return &taskObj.Task, nil
}

// RunTaskLoop runs a loop to process tasks, until the context is cancelled
// or an error occurs.
func (manager *Manager) RunTaskLoop(ctx context.Context, ch chan<- *schema.Task) error {
	delta := schema.TaskPeriod
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	// Loop until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			// Check for matured tasks
			task, err := manager.NextTask(ctx)
			if err != nil {
				return err
			}

			if task != nil {
				ch <- task
				delta = 100 * time.Millisecond
			} else {
				delta = schema.TaskPeriod
			}

			// Next loop
			timer.Reset(delta)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - NOTIFICATION

// RunNotificationLoop runs a loop to process database notifications, until the context is cancelled
// or an error occurs.
func (manager *Manager) RunNotificationLoop(ctx context.Context, ch chan<- *pg.Notification) error {
	// Subscribe to topics
	for _, topic := range manager.topics {
		if err := manager.listener.Listen(ctx, topic); err != nil {
			return err
		}
	}
	defer func() {
		for _, topic := range manager.topics {
			manager.listener.Unlisten(ctx, topic)
		}
	}()

	// Loop until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if notification, err := manager.listener.WaitForNotification(ctx); err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					return err
				}
			} else {
				ch <- notification
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}

func defaultWorker() (string, error) {
	if hostname, err := os.Hostname(); err != nil {
		return "", httpresponse.ErrInternalError.With(err)
	} else {
		return fmt.Sprint(hostname, ".", os.Getpid()), nil
	}
}
