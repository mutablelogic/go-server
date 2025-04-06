package client

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) CreateCert(ctx context.Context, meta schema.CertCreateMeta) (*schema.CertMeta, error) {
	req, err := client.NewJSONRequest(meta)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.CertMeta
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("cert")); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) GetCert(ctx context.Context, name string) (*schema.Cert, error) {
	req := client.NewRequest()

	// Perform request
	var response schema.Cert
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("cert", name)); err != nil {
		return nil, err
	}

	// Return the responses
	return &response, nil
}

func (c *Client) DeleteCert(ctx context.Context, name string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("cert", name))
}
