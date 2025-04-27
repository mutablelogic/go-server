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

func (c *Client) ListTokens(ctx context.Context, user string, opts ...Opt) (*schema.TokenList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.TokenList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("token", user), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateToken(ctx context.Context, user string, meta schema.TokenMeta) (*schema.Token, error) {
	// Make request
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Token
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("token", user)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetToken(ctx context.Context, user string, id uint64) (*schema.Token, error) {
	// Make request
	req := client.NewRequest()

	// Perform request
	var response schema.Token
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("token", user, id)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) DeleteToken(ctx context.Context, user string, id uint64, force bool) error {
	opts, err := applyOpts(WithForce(force))
	if err != nil {
		return err
	}
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("token", user, id), client.OptQuery(opts.Values))
}

func (c *Client) UpdateToken(ctx context.Context, user string, id uint64, meta schema.TokenMeta) (*schema.Token, error) {
	// Make request
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Token
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("token", user, id)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
