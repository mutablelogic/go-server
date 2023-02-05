package router

import "github.com/mutablelogic/go-server/pkg/client"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	c      *client.Client
	prefix string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClient(client *client.Client, prefix string) *Client {
	if client == nil || prefix == "" {
		return nil
	}
	return &Client{
		c:      client,
		prefix: prefix,
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return the prefix associated with this client
func (c *Client) Prefix() string {
	return c.prefix
}

// Return the list of services
func (c *Client) Services() ([]Gateway, error) {
	var response []Gateway
	if err := c.c.Do(nil, &response, client.OptPath(c.Prefix())); err != nil {
		return nil, err
	} else {
		return response, nil
	}
}

// Return routes and middleware for a service
func (c *Client) Routes(prefix string) (Gateway, error) {
	var response Gateway
	if err := c.c.Do(nil, &response, client.OptPath(c.Prefix(), prefix)); err != nil {
		return Gateway{}, err
	} else {
		return response, nil
	}
}
