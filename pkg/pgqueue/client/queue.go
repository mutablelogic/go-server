package client

import (
	"context"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListQueues(ctx context.Context, opts ...Opt) (*schema.QueueList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.QueueList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("queue"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) GetQueue(ctx context.Context, name string) (*schema.Queue, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Queue
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("queue", name)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) CreateQueue(ctx context.Context, meta schema.QueueMeta) (*schema.Queue, error) {
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Queue
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("queue")); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) DeleteQueue(ctx context.Context, name string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("queue", name))
}

func (c *Client) UpdateQueue(ctx context.Context, name string, meta schema.QueueMeta) (*schema.Queue, error) {
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Queue
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("queue", name)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

func (c *Client) CleanQueue(ctx context.Context, name string) ([]schema.Task, error) {
	req := client.NewRequest()

	// Perform request
	var response []schema.Task
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("queue", name, "clean")); err != nil {
		return nil, err
	}

	// Return the response
	return response, nil
}
