package nginx_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/nginx"
	"github.com/stretchr/testify/assert"
)

const (
	NginxAvailable = `../../etc/test/nginx`
)

func Test_Config_000(t *testing.T) {
	assert := assert.New(t)
	// No enabled folder
	config, err := nginx.NewConfig(NginxAvailable, "")
	assert.NoError(err)
	if err == nil {
		t.Log(config)
	}
}
