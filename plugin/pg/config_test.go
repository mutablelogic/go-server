package pg_test

import (
	"context"
	"testing"

	// Packages
	pg "github.com/mutablelogic/go-server/plugin/pg"
	assert "github.com/stretchr/testify/assert"
)

func Test_config_001(t *testing.T) {
	assert := assert.New(t)
	config := pg.Config{}
	assert.Equal("pgpool", config.Name())
	assert.NotEmpty(config.Description())
}

func Test_config_002(t *testing.T) {
	assert := assert.New(t)
	config := pg.Config{}
	task, err := config.New(context.TODO())
	if assert.NoError(err) {
		assert.NotNil(task)
	}
}
