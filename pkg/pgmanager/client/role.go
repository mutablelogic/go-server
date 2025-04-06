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

func (c *Client) ListRoles(ctx context.Context, opts ...Opt) (*schema.RoleList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.RoleList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("role"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateRole(ctx context.Context, role schema.RoleMeta) (*schema.Role, error) {
	req, err := client.NewJSONRequest(role)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Role
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("role")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetRole(ctx context.Context, name string) (*schema.Role, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Role
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("role", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) DeleteRole(ctx context.Context, name string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("role", name))
}

func (c *Client) UpdateRole(ctx context.Context, name string, meta schema.RoleMeta) (*schema.Role, error) {
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Role
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("role", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
