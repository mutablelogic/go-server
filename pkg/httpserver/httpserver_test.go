package httpserver_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/stretchr/testify/assert"
)

func Test_httpserver_001(t *testing.T) {
	assert := assert.New(t)
	config := httpserver.Config{}
	assert.NotEmpty(config.Name())
	assert.NotEmpty(config.Description())
}

func Test_httpserver_002(t *testing.T) {
	assert := assert.New(t)
	config := httpserver.Config{}
	server, err := config.New()
	assert.NoError(err)
	assert.NotNil(server)
	t.Log(server)
}
