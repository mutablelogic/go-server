package main_test

import (
	"context"
	"testing"

	// Packages
	server "github.com/mutablelogic/go-server"
	pg "github.com/mutablelogic/go-server/plugin/pg"
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
	pg, err := pg.Config{}.New(context.TODO())
	if assert.NoError(err) {
		assert.NotNil(pg)
		pgqueue, err := pgqueue.Config{Pool: pg.(server.PG)}.New(context.TODO())
		if assert.NoError(err) {
			assert.NotNil(pgqueue)
		}
	}
}
