package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
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

func (c *Client) GetResourceInstance(ctx context.Context, name string) (*schema.GetResourceInstanceResponse, error) {
	var response schema.GetResourceInstanceResponse
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("resource", name)); err != nil {
		return nil, err
	}
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

func (c *Client) DestroyResourceInstance(ctx context.Context, name string, cascade bool) (*schema.DestroyResourceInstanceResponse, error) {
	request, err := client.NewJSONRequestEx(http.MethodDelete, nil, "")
	if err != nil {
		return nil, err
	}

	opts := []client.RequestOpt{client.OptPath("resource", name)}
	if cascade {
		opts = append(opts, client.OptQuery(map[string][]string{"cascade": {"true"}}))
	}

	var response schema.DestroyResourceInstanceResponse
	if err := c.DoWithContext(ctx, request, &response, opts...); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetOpenAPI looks up the named router instance, reads its endpoint
// from state, and fetches {endpoint}/openapi.json.
func (c *Client) GetOpenAPI(ctx context.Context, routerName string) (json.RawMessage, error) {
	// Get the router instance to find its endpoint
	router, err := c.GetResourceInstance(ctx, routerName)
	if err != nil {
		return nil, fmt.Errorf("router %q: %w", routerName, err)
	}
	endpoint, ok := router.Instance.State["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, fmt.Errorf("router %q has no endpoint (is the server running?)", routerName)
	}

	// Fetch the OpenAPI spec from the router's endpoint
	resp, err := http.Get(endpoint + "/openapi.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openapi: %s", resp.Status)
	}

	var spec json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		return nil, err
	}
	return spec, nil
}
