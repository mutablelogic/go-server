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

func Test_reflect_001(t *testing.T) {
	assert := assert.New(t)

	meta, err := provider.NewPluginMeta(httpserver.Config{})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	plugin := &httpserver.Config{}
	assert.NoError(meta.Set(plugin, "listen", "value"))
	assert.Error(meta.Set(plugin, "listen", 99))
	assert.NoError(meta.Set(plugin, "tls.key", "key"))
	assert.NoError(meta.Set(plugin, "timeout", nil))

	t.Log(plugin)

}

func Test_reflect_002(t *testing.T) {
	assert := assert.New(t)

	meta, err := provider.NewPluginMeta(router.Config{})
	if !assert.NoError(err) {
		t.SkipNow()
	}
	t.Log(meta)
}

func Test_reflect_003(t *testing.T) {
	assert := assert.New(t)

	meta, err := provider.NewPluginMeta(logger.Config{})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	plugin := &logger.Config{}
	assert.NoError(meta.Set(plugin, "flags", []string{"0", "1"}))
	assert.NoError(meta.Set(plugin, "flags", []string{}))

	t.Log(plugin)
}
