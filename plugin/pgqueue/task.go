package pgqueue

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
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	sync.WaitGroup
	client  *pgqueue.Client
	tickers map[string]exec
}

type exec struct {
	deadline time.Duration
	fn       server.PGCallback
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func taskWith(queue *pgqueue.Client) *task {
	return &task{
		client:  queue,
		tickers: make(map[string]exec, 10),
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

	// Emit ticker events
	t.Add(1)
	go func(ctx context.Context) {
		defer t.Done()
		result = errors.Join(result, t.client.RunTickerLoop(ctx, tickerch))
	}(ctx)

	// Continue until context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case ticker := <-tickerch:
			t.execTicker(ctx, ticker)
		}
	}

	// Wait for goroutines to finish
	t.Wait()

	// Return any errors
	return result
}

// Register a ticker with a callback, and return the registered ticker
func (t *task) RegisterTicker(ctx context.Context, meta schema.TickerMeta, fn server.PGCallback) error {
	ticker, err := t.client.RegisterTicker(ctx, meta)
	if err != nil {
		return err
	}

	// Add the ticker to the map
	t.tickers[ticker.Ticker] = exec{
		deadline: types.PtrDuration(ticker.Interval),
		fn:       fn,
	}

	// Return success
	return nil
}

// Register a queue with a callback, and return the registered queue
func (t *task) RegisterQueue(ctx context.Context, meta schema.Queue, fn server.PGCallback) error {
	_, err := t.client.RegisterQueue(ctx, meta)
	if err != nil {
		return err
	}

	// Return success
	return nil
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
func (t *task) exec(ctx context.Context, fn exec, in any) error {
	var result error

	// Create a context with a deadline
	deadline, cancel := contextWithDeadline(ctx, fn.deadline)
	defer cancel()

	// Catch panics
	defer func() {
		if r := recover(); r != nil {
			result = errors.Join(result, fmt.Errorf("PANIC: %v", r))
		}
	}()

	// Run the ticker handler
	err := fn.fn(deadline, in)
	if err != nil {
		result = errors.Join(result, err)
	}

	// Concatenate any errors from the deadline
	if deadline.Err() != nil {
		result = errors.Join(result, deadline.Err())
	}

	// Return any errors
	return result
}

// Execute a ticker callback
func (t *task) execTicker(ctx context.Context, ticker *schema.Ticker) {
	fn, exists := t.tickers[ticker.Ticker]
	if !exists {
		return
	}

	// Execute the callback function in a goroutine
	t.Add(1)
	go func(ctx context.Context) {
		defer t.Done()
		if err := t.exec(ctx, fn, ticker); err != nil {
			// TODO: Emit any errors on an err channel
			log.Println(ctx, "TICKER", ticker.Ticker, err)
		}
	}(contextWithTicker(ctx, ticker))
}

/*

	// Defer closing the listener
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		t.listener.Close(ctx)
	}()

	// Subscribe to pg topics
	for _, topic := range t.topics {
		if err := t.listener.Listen(ctx, topic); err != nil {
			return err
		}
	}

	// Create channel for notifications, which we close when we're done
	evt := make(chan *pg.Notification)
	defer close(evt)

	// Notifications
	t.wg.Add(1)
	go func(ctx context.Context) {
		defer t.wg.Done()
		runNotificationsForListener(ctx, t.listener, evt)
	}(ctx)

	// Ticker namespaces
	tickerch := make(chan *schema.Ticker)
	defer close(tickerch)
	for namespace := range t.ticker {
		t.wg.Add(1)
		go func(ctx context.Context) {
			defer t.wg.Done()
			t.runTickerForNamespace(ctx, namespace, tickerch)
		}(ctx)
	}

	// Queue processing
	defer close(t.taskch)

	// Process notifications and tickers in the main loop
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case notification := <-evt:
			// Notification that a queue task has been inserted
			now := time.Now()
			provider.Log(ctx).With("channel", notification.Channel, "payload", string(notification.Payload)).Print(ctx, "NOTIFICATION")
			if err := t.processNotificationTopic(ctx, notification.Channel, string(notification.Payload)); err != nil {
				provider.Log(ctx).With("channel", notification.Channel, "duration_ms", time.Since(now).Milliseconds()).Print(ctx, "  ", err)
			} else {
				provider.Log(ctx).With("channel", notification.Channel, "duration_ms", time.Since(now).Milliseconds()).Print(ctx, "  ", "COMPLETED")
			}
		case ticker := <-tickerch:
			// Notification that a ticker has fired
			now := time.Now()
			provider.Log(ctx).With("namespace", ticker.Namespace, "ticker", ticker.Ticker).Print(ctx, "TICKER")
			if err := t.processTicker(ctx, ticker); err != nil {
				provider.Log(ctx).With("namespace", ticker.Namespace, "ticker", ticker.Ticker, "duration_ms", time.Since(now).Milliseconds()).Print(ctx, "  ", err)
			} else {
				provider.Log(ctx).With("namespace", ticker.Namespace, "ticker", ticker.Ticker, "duration_ms", time.Since(now).Milliseconds()).Print(ctx, "  ", "COMPLETED")
			}
		case task := <-t.taskch:
			// Notification that a task has been locked for processing
			now := time.Now()
			provider.Log(ctx).With("queue", task.Queue, "id", *task.Id).Print(ctx, "TASK")
			if result, err := t.processTask(ctx, task); err != nil {
				provider.Log(ctx).With("queue", task.Queue, "id", *task.Id, "duration_ms", time.Since(now).Milliseconds()).Print(ctx, "  ", err)
			} else {
				provider.Log(ctx).With("queue", task.Queue, "id", *task.Id, "result", result, "duration_ms", time.Since(now).Milliseconds()).Print(ctx, "  ", "COMPLETED")
			}
		}
	}

	// Wait for goroutines to finish
	t.wg.Wait()

	// Return success
	return nil
}*/
