package pgqueue

import (
	"context"
	"errors"
	"fmt"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	conn     pg.PoolConn
	listener pg.Listener
	topics   []string
	worker   string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*Client, error) {
	self := new(Client)
	opts, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	} else {
		self.worker = opts.worker
	}

	// Create a listener
	if listener := pg.NewListener(conn); listener == nil {
		return nil, httpresponse.ErrInternalError.Withf("Cannot create listener")
	} else {
		self.listener = listener
		self.conn = conn.With(
			"schema", schema.SchemaName,
			"ns", opts.namespace,
		).(pg.PoolConn)
		self.topics = []string{schema.TopicQueueInsert}
	}

	// If the schema does not exist, then bootstrap it
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		if exists, err := pg.SchemaExists(ctx, conn, schema.SchemaName); err != nil {
			return err
		} else if !exists {
			return schema.Bootstrap(ctx, conn)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// CreateQueue creates a new queue, and returns it.
func (client *Client) CreateQueue(ctx context.Context, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		return client.conn.Insert(ctx, &queue, meta)
	}); err != nil {
		return nil, err
	}
	return &queue, nil
}

// GetQueue returns a queue with the given name.
func (client *Client) GetQueue(ctx context.Context, name string) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Get(ctx, &queue, schema.QueueName(name)); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Queue %q not found", name)
		}
		return nil, err
	}
	return &queue, nil
}

// DeleteQueue deletes a queue with the given name, and returns the deleted queue.
func (client *Client) DeleteQueue(ctx context.Context, name string) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Delete(ctx, &queue, schema.QueueName(name)); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Queue %q not found", name)
		}
		return nil, err
	}
	return &queue, nil
}

// UpdateQueue updates an existing queue with the given name, and returns the queue.
func (client *Client) UpdateQueue(ctx context.Context, name string, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Update(ctx, &queue, schema.QueueName(name), meta); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Queue %q not found", name)
		}
		return nil, err
	}
	return &queue, nil
}

// ListQueues returns all queues as a list
func (client *Client) ListQueues(ctx context.Context, opt ...Opt) (*schema.QueueList, error) {
	var list schema.QueueList

	// Apply options
	opts, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}

	// Perform list
	list.Body = make([]schema.Queue, 0, 10)
	if err := client.conn.List(ctx, &list, schema.QueueListRequest{
		OffsetLimit: opts.OffsetLimit,
	}); err != nil {
		return nil, err
	}
	return &list, nil
}

// CreateTicker creates a new ticker, and returns it.
func (client *Client) CreateTicker(ctx context.Context, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		return client.conn.Insert(ctx, &ticker, meta)
	}); err != nil {
		return nil, err
	}
	fmt.Println("Created ticker:", ticker)
	return &ticker, nil
}

// GetTicker returns a ticker with the given name.
func (client *Client) GetTicker(ctx context.Context, name string) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Get(ctx, &ticker, schema.TickerName(name)); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Ticker %q not found", name)
		}
		return nil, err
	}
	return &ticker, nil
}
