/*
Implements an API client for the nginx API (https://github.com/mutablelogic/go-server/pkg/handler/nginx)
*/
package client

import (
	"time"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new API client, providing the endpoint (ie, http://example.com/api/nginx)
func New(endPoint string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endPoint))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return health of the API - nginx version, uptime
func (c *Client) Health() (string, time.Duration, error) {
	var response struct {
		Version string `json:"version"`
		Uptime  int64  `json:"uptime"`
	}
	if err := c.Do(nil, &response); err != nil {
		return "", 0, err
	}
	return response.Version, time.Duration(response.Uptime) * time.Second, nil
}

// Reload the nginx configuration
func (c *Client) Reload() error {
	return c.Do(client.MethodPut, nil, client.OptPath("reload"))
}

// Reopen the log files (after rotation)
func (c *Client) Reopen() error {
	return c.Do(client.MethodPut, nil, client.OptPath("reopen"))
}

// Test the configuration
func (c *Client) Test() error {
	return c.Do(client.MethodPut, nil, client.OptPath("test"))
}
