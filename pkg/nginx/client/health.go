package client

import (
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type ReqHealth struct {
	client.Payload
}

type RespHealth struct {
}

func (req ReqHealth) Method() string {
	return http.MethodGet
}

func (req ReqHealth) Accept() string {
	return client.ContentTypeJson
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Health() (string, error) {
	var response RespHealth
	if err := c.Do(ReqHealth{}, &response, client.OptPath("/")); err != nil {
		return "", err
	} else {
		return "OK", nil
	}
}
