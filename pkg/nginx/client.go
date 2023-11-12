package nginx

import (
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
	Prefix string
}

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type reqpost struct{}

func (r reqpost) Method() string {
	return http.MethodPost
}

func (r reqpost) Accept() string {
	return ""
}

func (r reqpost) Type() string {
	return ""
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClient(client *client.Client, prefix string) *Client {
	if client == nil || prefix == "" {
		return nil
	}
	return &Client{client, prefix}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return the version and uptime for nginx
func (c *Client) Health() (HealthResponse, error) {
	var response HealthResponse
	if err := c.Do(nil, &response, client.OptPath(c.Prefix)); err != nil {
		return response, err
	} else {
		return response, nil
	}
}

// Test nginx configuration, return error if test was not successful
func (c *Client) Test() error {
	return c.Do(reqpost{}, nil, client.OptPath(c.Prefix, "test"))
}

// Reload nginx configuration, return error if not successful
func (c *Client) Reload() error {
	return c.Do(reqpost{}, nil, client.OptPath(c.Prefix, "reload"))
}

// Reopen nginx log files
func (c *Client) Reopen() error {
	return c.Do(reqpost{}, nil, client.OptPath(c.Prefix, "reopen"))
}
