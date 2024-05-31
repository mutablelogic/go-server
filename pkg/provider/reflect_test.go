package provider_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/router"
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/mutablelogic/go-server/pkg/middleware/logger"
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_reflect_001(t *testing.T) {
	assert := assert.New(t)

	meta, err := provider.NewMeta(httpserver.Config{})
	assert.NoError(err)
	for k, v := range meta.Fields {
		t.Log(meta.Name, k, v)
	}

	meta, err = provider.NewMeta(router.Config{})
	assert.NoError(err)
	for k, v := range meta.Fields {
		t.Log(meta.Name, k, v)
	}

	meta, err = provider.NewMeta(logger.Config{})
	assert.NoError(err)
	for k, v := range meta.Fields {
		t.Log(meta.Name, k, v)
	}
}
