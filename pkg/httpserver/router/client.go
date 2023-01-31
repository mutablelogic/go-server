package router

import "github.com/mutablelogic/go-server/pkg/client"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type routerclient struct {
	c      *client.Client
	prefix string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClient(client *client.Client, prefix string) *routerclient {
	if client == nil || prefix == "" {
		return nil
	}
	return &routerclient{
		c:      client,
		prefix: prefix,
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return the prefix associated with this client
func (c *routerclient) Prefix() string {
	return c.prefix
}

// Return the list of services
func (c *routerclient) Services() ([]Gateway, error) {
	var response []Gateway
	if err := c.c.Do(nil, &response, client.OptPath(c.Prefix())); err != nil {
		return nil, err
	} else {
		return response, nil
	}
}

// Return routes and middleware for a service
func (c *routerclient) Routes(prefix string) (Gateway, error) {
	var response Gateway
	if err := c.c.Do(nil, &response, client.OptPath(c.Prefix(), prefix)); err != nil {
		return Gateway{}, err
	} else {
		return response, nil
	}
}
