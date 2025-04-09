package client

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListSchemas(ctx context.Context, opts ...Opt) (*schema.SchemaList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.SchemaList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("schema"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetSchema(ctx context.Context, name string) (*schema.Database, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Database
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("schema", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateSchema(ctx context.Context, meta schema.SchemaMeta) (*schema.Schema, error) {
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Schema
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("schema")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
