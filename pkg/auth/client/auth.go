package client

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Auth(ctx context.Context, value string) (*schema.User, error) {
	req, err := client.NewJSONRequest(value)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.User
	if err := c.DoWithContext(ctx, req, &response); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
