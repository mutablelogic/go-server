package client

import (
	"context"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Auth(ctx context.Context, dn, password string) (*schema.Object, error) {
	req, err := client.NewJSONRequest(password)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Object
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("auth", dn)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) ChangePassword(ctx context.Context, dn, password string) (*schema.Object, error) {
	req, err := client.NewJSONRequestEx(http.MethodPut, password, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	// TODO: Retrieve the new password from the response
	var response schema.Object
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("auth", dn)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
