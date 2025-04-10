package client

import (
	"context"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListDatabases(ctx context.Context, opts ...Opt) (*schema.DatabaseList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.DatabaseList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("database"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetDatabase(ctx context.Context, name string) (*schema.Database, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Database
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("database", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateDatabase(ctx context.Context, database schema.Database) (*schema.Database, error) {
	req, err := client.NewJSONRequest(database)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Database
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("database")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, name string, opt ...Opt) error {
	opts, err := applyOpts(opt...)
	if err != nil {
		return err
	}
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("database", name), client.OptQuery(opts.Values))
}

func (c *Client) UpdateDatabase(ctx context.Context, name string, meta schema.Database) (*schema.Database, error) {
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Database
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("database", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
