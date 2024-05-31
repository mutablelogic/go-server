package client_test

import (
	"os"
	"testing"

	// Packages
	client "github.com/mutablelogic/go-client"
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx/client"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := nginx.New(GetEndpoint(t), client.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	version, uptime, err := client.Health()
	assert.NoError(err)
	assert.NotEmpty(version)
	assert.NotZero(uptime)
	t.Log("version=", version)
	t.Log("uptime=", uptime)
}

func Test_client_002(t *testing.T) {
	assert := assert.New(t)
	client, err := nginx.New(GetEndpoint(t), client.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NoError(client.Test())
}

func Test_client_003(t *testing.T) {
	assert := assert.New(t)
	client, err := nginx.New(GetEndpoint(t), client.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NoError(client.Reopen())
}

func Test_client_004(t *testing.T) {
	assert := assert.New(t)
	client, err := nginx.New(GetEndpoint(t), client.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NoError(client.Reload())
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetEndpoint(t *testing.T) string {
	key := os.Getenv("NGINX_ENDPOINT")
	if key == "" {
		t.Skip("NGINX_ENDPOINT not set")
		t.SkipNow()
	}
	return key
}
