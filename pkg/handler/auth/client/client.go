/*
Implements an API client for the Token auth API (https://github.com/mutablelogic/go-server/pkg/handler/auth)
*/
package client

import (

	// Packages
	"time"

	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-server/pkg/handler/auth"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new API client, providing the endpoint (ie, http://example.com/api/auth)
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

// Return all tokens
func (c *Client) List() ([]auth.Token, error) {
	var response []auth.Token
	if err := c.Do(nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// Get token details (apart from the value)
func (c *Client) Get(name string) (auth.Token, error) {
	var response auth.Token
	if err := c.Do(nil, &response, client.OptPath(name)); err != nil {
		return auth.Token{}, err
	}
	return response, nil
}

// Delete a token
func (c *Client) Delete(name string) error {
	if err := c.Do(client.MethodDelete, nil, client.OptPath(name)); err != nil {
		return err
	}
	// Return success
	return nil
}

// Create a token with name, duration and scopes, and return the token
func (c *Client) Create(name string, expires_in time.Duration, scopes ...string) (auth.Token, error) {
	var response auth.Token

	// Request->Response
	if payload, err := client.NewJSONRequest(auth.NewCreateToken(name, expires_in, scopes...)); err != nil {
		return auth.Token{}, err
	} else if err := c.Do(payload, &response); err != nil {
		return auth.Token{}, err
	}

	// Return success
	return response, nil
}
