package httpclient

import (
	"context"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListResources(ctx context.Context, req schema.ListResourcesRequest) (*schema.ListResourcesResponse, error) {
	// Set request options
	opts := []client.RequestOpt{
		client.OptPath("resource"),
		client.OptQuery(req.Query()),
	}

	// Perform GET request
	var response schema.ListResourcesResponse
	if err := c.DoWithContext(ctx, nil, &response, opts...); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}

func (c *Client) CreateResourceInstance(ctx context.Context, req schema.CreateResourceInstanceRequest) (*schema.CreateResourceInstanceResponse, error) {
	request, err := client.NewJSONRequest(req)
	if err != nil {
		return nil, err
	}

	// Perform POST request
	var response schema.CreateResourceInstanceResponse
	if err := c.DoWithContext(ctx, request, &response, client.OptPath("resource")); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}

func (c *Client) UpdateResourceInstance(ctx context.Context, name string, req schema.UpdateResourceInstanceRequest) (*schema.UpdateResourceInstanceResponse, error) {
	request, err := client.NewJSONRequestEx(http.MethodPatch, req, "")
	if err != nil {
		return nil, err
	}

	// Perform PATCH request
	var response schema.UpdateResourceInstanceResponse
	if err := c.DoWithContext(ctx, request, &response, client.OptPath("resource", name)); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}
