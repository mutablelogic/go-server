package client

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListUsers(ctx context.Context, opts ...Opt) (*schema.ObjectList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.ObjectList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("user"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateUser(ctx context.Context, meta schema.Object) (*schema.Object, error) {
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Object
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("user")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetUser(ctx context.Context, user string) (*schema.Object, error) {
	var resp schema.Object

	// Perform request
	if err := c.DoWithContext(ctx, client.MethodGet, &resp, client.OptPath("user", user)); err != nil {
		return nil, err
	}

	// Return the response
	return &resp, nil
}

func (c *Client) DeleteUser(ctx context.Context, user string) error {
	// Perform request
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("user", user))
}
