package provider_test

import (
	"testing"

	// Packages
	provider "github.com/mutablelogic/go-server/pkg/provider"
	httpserver "github.com/mutablelogic/go-server/plugin/httpserver"
	assert "github.com/stretchr/testify/assert"
)

func Test_Provider_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		provider, err := provider.New(nil)
		if assert.NoError(err) {
			assert.NotNil(provider)
		}
	})

	t.Run("2", func(t *testing.T) {
		provider, err := provider.New(nil, httpserver.Config{})
		if assert.NoError(err) {
			assert.NotNil(provider)
		}
	})
}
