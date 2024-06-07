package provider_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/logger"
	"github.com/mutablelogic/go-server/pkg/handler/router"
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_provider_001(t *testing.T) {
	assert := assert.New(t)

	provider, err := provider.New(httpserver.Config{}, router.Config{}, logger.Config{})
	assert.NoError(err)

	t.Log(provider)
}
