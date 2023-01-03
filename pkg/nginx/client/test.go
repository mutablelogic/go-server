package client

import (
	// Packages
	"net/http"

	client "github.com/mutablelogic/go-server/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type ReqPost struct{}

func (r ReqPost) Method() string {
	return http.MethodPost
}

func (r ReqPost) Accept() string {
	return ""
}

func (r ReqPost) Type() string {
	return ""
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Test() error {
	return c.Do(ReqPost{}, nil, client.OptPath("/test"))
}

func (c *Client) Reload() error {
	return c.Do(ReqPost{}, nil, client.OptPath("/reload"))
}

func (c *Client) Reopen() error {
	return c.Do(ReqPost{}, nil, client.OptPath("/reopen"))
}
