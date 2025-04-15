package client

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListUsers(ctx context.Context, opts ...Opt) (*schema.UserListResponse, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.UserListResponse
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("user"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateUser(ctx context.Context, meta schema.UserMeta) (*schema.User, error) {
	// Make request
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.User
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("user")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
