package client

import (
	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// NEW

func New(endpoint string) (*Client, error) {
	// Create client
	client, err := client.New(client.OptEndpoint(endpoint))
	if err != nil {
		return nil, err
	}

	// Return success
	return &Client{client}, nil
}
