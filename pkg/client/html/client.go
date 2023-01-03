/*
html implements a generic API client which parses HTML and XML files. Mostly used
to test the client package.
*/
package client

import (
	"io"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/client"
	"golang.org/x/net/html"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Request struct {
	client.Payload `json:"-"`
}

type Response string

func (r Request) Method() string {
	return http.MethodGet
}

func (r Request) Type() string {
	// Return empty mimetype for GET requests
	return ""
}

func (r Request) Accept() string {
	return client.ContentTypeTextHTML
}

func (response *Response) Unmarshal(mimetype string, r io.Reader) error {
	parser := html.NewTokenizer(r)
	for {
		token := parser.Next()
		switch token {
		case html.ErrorToken:
			if parser.Err() == io.EOF {
				return nil
			} else {
				return parser.Err()
			}
		case html.TextToken:
			*response += Response(parser.Text())
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(endpoint string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endpoint))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Get(path string) (string, error) {
	var response Response
	if err := c.Do(nil, &response, client.OptPath(path)); err != nil {
		return "", err
	}
	return string(response), nil
}
