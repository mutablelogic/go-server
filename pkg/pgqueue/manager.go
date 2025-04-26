package pgqueue

import (
	"context"
	"errors"
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
	conn pg.PoolConn
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*Manager, error) {
	self := new(Manager)
	opts, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}

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

	// Return success
	return self, nil
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

// RunTickerLoop runs a loop to process matured tickers, until the context is cancelled.
func (manager *Manager) RunTickerLoop(ctx context.Context, ch chan<- *schema.Ticker) error {
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
			ticker, err := manager.NextTicker(ctx)
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
func (manager *Manager) RegisterQueue(ctx context.Context, meta schema.Queue) (*schema.Queue, error) {
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}
