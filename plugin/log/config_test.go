package log_test

import (
	"testing"

	// Packages
	log "github.com/mutablelogic/go-server/plugin/log"
	assert "github.com/stretchr/testify/assert"
)

func Test_Config_001(t *testing.T) {
	assert := assert.New(t)
	config := log.Config{}
	assert.Equal("log", config.Name())
	assert.NotEmpty(config.Description())
}
