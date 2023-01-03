package client

import (
	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Health struct {
	Version    string `json:"version"`
	UptimeSecs uint64 `json:"uptime_secs"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Health() (Health, error) {
	var response Health
	if err := c.Do(nil, &response, client.OptPath("/")); err != nil {
		return response, err
	} else {
		return response, nil
	}
}
