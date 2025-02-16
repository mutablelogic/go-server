package httpserver_test

import (
	"crypto/tls"
	"testing"

	// Packages
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	assert "github.com/stretchr/testify/assert"
)

func Test_Server_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		server, err := httpserver.New(":8080", nil, nil)
		if assert.NoError(err) {
			assert.NotNil(server)
		}
	})
	t.Run("2", func(t *testing.T) {
		server, err := httpserver.New("", nil, nil)
		if assert.NoError(err) {
			assert.NotNil(server)
		}
	})
	t.Run("3", func(t *testing.T) {
		server, err := httpserver.New("", nil, &tls.Config{})
		if assert.NoError(err) {
			assert.NotNil(server)
		}
	})
}
