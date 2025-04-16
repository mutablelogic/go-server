package client

import (
	"context"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListUsers(ctx context.Context, opts ...Opt) (*schema.UserList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.UserList
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

func (c *Client) GetUser(ctx context.Context, name string) (*schema.User, error) {
	// Make request
	req := client.NewRequest()

	// Perform request
	var response schema.User
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("user", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) DeleteUser(ctx context.Context, name string, force bool) error {
	opts, err := applyOpts(WithForce(force))
	if err != nil {
		return err
	}
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("user", name), client.OptQuery(opts.Values))
}

func (c *Client) UpdateUser(ctx context.Context, name string, meta schema.UserMeta) (*schema.User, error) {
	// Make request
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.User
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("user", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
