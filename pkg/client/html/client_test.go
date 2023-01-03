package client_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-server/pkg/client"
	html "github.com/mutablelogic/go-server/pkg/client/html"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	// Fetch a news page. The response contains only the text in the HTML file
	assert := assert.New(t)
	client, err := html.New("http://news.bbc.co.uk/", opts.OptTrace(os.Stderr, false))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.Get("/")
	assert.NoError(err)
	assert.NotEmpty(response)
}
