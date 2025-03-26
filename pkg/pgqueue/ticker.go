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
// PUBLIC METHODS

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
