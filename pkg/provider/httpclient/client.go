package httpclient

import (
	// Packages
	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a HTTP client that wraps the base HTTP client
// and provides typed methods for interacting with the provider API.
type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new provider HTTP client with the given base URL and options.
// The url parameter should point to the provider API endpoint, e.g.
// "http://localhost:8080/api/provider".
func New(url string, opts ...client.ClientOpt) (*Client, error) {
	c := new(Client)
	if client, err := client.New(append(opts, client.OptEndpoint(url))...); err != nil {
		return nil, err
	} else {
		c.Client = client
	}
	return c, nil
}
