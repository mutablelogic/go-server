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

func (c *Client) ListTablespaces(ctx context.Context, opts ...Opt) (*schema.TablespaceList, error) {
	req := client.NewRequest()

	// Apply options
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.TablespaceList
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("tablespace"), client.OptQuery(opt.Values)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) CreateTablespace(ctx context.Context, meta schema.TablespaceMeta, location string) (*schema.Tablespace, error) {
	type create struct {
		schema.TablespaceMeta
		Location string `json:"location"`
	}
	req, err := client.NewJSONRequest(create{meta, location})
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Tablespace
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("tablespace")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetTablespace(ctx context.Context, name string) (*schema.Tablespace, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Tablespace
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("tablespace", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) DeleteTablespace(ctx context.Context, name string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("tablespace", name))
}

func (c *Client) UpdateTablespace(ctx context.Context, name string, meta schema.TablespaceMeta) (*schema.Tablespace, error) {
	req, err := client.NewJSONRequestEx(http.MethodPatch, meta, "")
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.Tablespace
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("tablespace", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}
