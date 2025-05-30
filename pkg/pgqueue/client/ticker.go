package client

import (
	"context"
	"fmt"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListTickers(ctx context.Context, opts ...Opt) (*schema.TickerList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.TickerList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("ticker"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) GetTicker(ctx context.Context, name string) (*schema.Ticker, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Ticker
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("ticker", name)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) CreateTicker(ctx context.Context, meta schema.TickerMeta) (*schema.Ticker, error) {
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Ticker
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("ticker")); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) DeleteTicker(ctx context.Context, name string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("ticker", name))
}

func (c *Client) UpdateTicker(ctx context.Context, name string, meta schema.TickerMeta) (*schema.Ticker, error) {
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Ticker
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("ticker", name)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) NextTicker(ctx context.Context) error {
	req, err := client.NewJSONRequestEx(http.MethodGet, nil, types.ContentTypeTextStream)
	if err != nil {
		return err
	}

	// Perform request
	var response schema.Ticker
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("ticker"), client.OptNoTimeout(), client.OptTextStreamCallback(func(event client.TextStreamEvent) error {
		fmt.Println("event", event)
		return nil
	})); err != nil {
		return err
	}

	// Return the response
	return nil
}
