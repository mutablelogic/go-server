package pgqueue_test

import (
	"context"
	"testing"

	// Packages
	pgqueue "github.com/mutablelogic/go-server/plugin/pgqueue"
	assert "github.com/stretchr/testify/assert"
)

func Test_config_001(t *testing.T) {
	assert := assert.New(t)
	config := pgqueue.Config{}
	assert.Equal("pgqueue", config.Name())
	assert.NotEmpty(config.Description())
}

func Test_config_002(t *testing.T) {
	assert := assert.New(t)
	config := pgqueue.Config{}
	task, err := config.New(context.TODO())
	if assert.NoError(err) {
		assert.NotNil(task)
	}
}
