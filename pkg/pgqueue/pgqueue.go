package pgqueue

import (
	"context"
	"errors"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	"github.com/mutablelogic/go-server/pkg/types"
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

func (client *Client) Close(ctx context.Context) error {
	var result error
	if client.listener != nil {
		result = errors.Join(result, client.listener.Close(ctx))
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Worker returns the worker name
func (client *Client) Worker() string {
	return client.worker
}

// RegisterQueue creates a new queue, or updates an existing queue, and returns it.
func (client *Client) RegisterQueue(ctx context.Context, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get a queue
		if err := conn.Get(ctx, &queue, schema.QueueName(meta.Queue)); err != nil && !errors.Is(err, pg.ErrNotFound) {
			return err
		} else if errors.Is(err, pg.ErrNotFound) {
			// If the queue does not exist, then create it
			return conn.Insert(ctx, &queue, meta)
		} else {
			// If the queue exists, then update it
			return conn.Update(ctx, &queue, schema.QueueName(meta.Queue), meta)
		}
	}); err != nil {
		return nil, err
	}
	return &queue, nil
}

// CreateQueue creates a new queue, and returns it.
func (client *Client) CreateQueue(ctx context.Context, meta schema.Queue) (*schema.Queue, error) {
	var queue schema.Queue
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := client.conn.Insert(ctx, &queue, meta); err != nil {
			return err
		} else if err := conn.Update(ctx, &queue, schema.QueueName(queue.Queue), meta); err != nil {
			return err
		}
		// Commit the transaction
		return nil
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
func (client *Client) ListQueues(ctx context.Context, req schema.QueueListRequest) (*schema.QueueList, error) {
	var list schema.QueueList

	// Perform list
	list.Body = make([]schema.Queue, 0, 10)
	if err := client.conn.List(ctx, &list, req); err != nil {
		return nil, err
	}
	return &list, nil
}

// RegisterTicker creates a new ticker, or updates an existing ticker, and returns it.
func (client *Client) RegisterTicker(ctx context.Context, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get a ticker
		if err := conn.Get(ctx, &ticker, schema.TickerName(meta.Ticker)); err != nil && !errors.Is(err, pg.ErrNotFound) {
			return err
		} else if errors.Is(err, pg.ErrNotFound) {
			// If the ticker does not exist, then create it
			return conn.Insert(ctx, &ticker, meta)
		} else {
			// If the ticker exists, then update it
			return conn.Update(ctx, &ticker, schema.TickerName(meta.Ticker), meta)
		}
	}); err != nil {
		return nil, err
	}
	return &ticker, nil
}

// CreateTicker creates a new ticker, and returns it.
func (client *Client) CreateTicker(ctx context.Context, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Tx(ctx, func(conn pg.Conn) error {
		return client.conn.Insert(ctx, &ticker, meta)
	}); err != nil {
		return nil, err
	}
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

// UpdateTicker updates a ticker with the given name.
func (client *Client) UpdateTicker(ctx context.Context, name string, meta schema.TickerMeta) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Update(ctx, &ticker, schema.TickerName(name), meta); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Ticker %q not found", name)
		}
		return nil, err
	}
	return &ticker, nil
}

// DeleteTicker deletes an existing ticker, and returns the deleted ticker.
func (client *Client) DeleteTicker(ctx context.Context, name string) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Delete(ctx, &ticker, schema.TickerName(name)); err != nil {
		if errors.Is(err, pg.ErrNotFound) {
			return nil, httpresponse.ErrNotFound.Withf("Ticker %q not found", name)
		}
		return nil, err
	}
	return &ticker, nil
}

// ListTickers returns all tickers in a namespace as a list
func (client *Client) ListTickers(ctx context.Context, req schema.TickerListRequest) (*schema.TickerList, error) {
	var list schema.TickerList

	// Perform list
	list.Body = make([]schema.Ticker, 0, 10)
	if err := client.conn.List(ctx, &list, req); err != nil {
		return nil, err
	}
	return &list, nil
}

// NextTicker returns the next matured ticker, or nil
func (client *Client) NextTicker(ctx context.Context) (*schema.Ticker, error) {
	var ticker schema.Ticker
	if err := client.conn.Get(ctx, &ticker, schema.TickerNext{}); errors.Is(err, pg.ErrNotFound) {
		// No matured ticker
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Return matured ticker
	return &ticker, nil
}

// RunTickerLoop runs a loop to process matured tickers, or NextTicker returns an error
func (client *Client) RunTickerLoop(ctx context.Context, ch chan<- *schema.Ticker) error {
	delta := schema.TickerPeriod
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	// Loop until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			// Check for matured tickers
			ticker, err := client.NextTicker(ctx)
			if err != nil {
				return err
			}

			if ticker != nil {
				ch <- ticker

				// Reset timer to minimum period
				if dur := types.PtrDuration(ticker.Interval); dur >= time.Second && dur < delta {
					delta = dur
				}
			}

			// Next loop
			timer.Reset(delta)
		}
	}
}
