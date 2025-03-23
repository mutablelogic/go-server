package pgqueue

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	pgqueue *pgqueue.Client
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func taskWith(queue *pgqueue.Client) *task {
	return &task{queue}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *task) Run(ctx context.Context) error {
	var result error
	var wg sync.WaitGroup

	// Defer closing the listener
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result = errors.Join(result, t.pgqueue.Close(ctx))
	}()

	// Create a ticker channel and defer close it
	tickerch := make(chan *schema.Ticker)
	defer close(tickerch)

	// Emit ticker events
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		result = errors.Join(result, t.pgqueue.RunTickerLoop(ctx, tickerch))
	}(ctx)

	// Continue until context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case ticker := <-tickerch:
			log.Println(ctx, "TICKER", ticker.Ticker)
		}
	}

	// Wait for goroutines to finish
	wg.Wait()

	// Return any errors
	return result
}

// Register a ticker with a callback, and return the registered ticker
func (t *task) RegisterTicker(ctx context.Context, meta schema.TickerMeta, fn server.PGCallback) error {
	_, err := t.pgqueue.RegisterTicker(ctx, meta)
	if err != nil {
		return err
	}

	// Return success
	return nil
}

// Register a queue with a callback, and return the registered queue
func (t *task) RegisterQueue(ctx context.Context, meta schema.Queue, fn server.PGCallback) error {
	_, err := t.pgqueue.RegisterQueue(ctx, meta)
	if err != nil {
		return err
	}

	// Return success
	return nil
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
