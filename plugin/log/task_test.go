package log_test

import (
	"context"
	"errors"
	"testing"

	// Packages
	server "github.com/mutablelogic/go-server"
	log "github.com/mutablelogic/go-server/plugin/log"
	assert "github.com/stretchr/testify/assert"
)

func Test_Task(t *testing.T) {
	assert := assert.New(t)
	log, err := log.Config{
		Debug: true,
	}.New(context.TODO())
	if assert.NoError(err) {
		assert.NotNil(log)
	}

	t.Run("Info", func(t *testing.T) {
		log.(server.Logger).Print(context.TODO(), "Hello world")
		log.(server.Logger).Printf(context.TODO(), "Hello %s", "world")
	})

	t.Run("Debug", func(t *testing.T) {
		log.(server.Logger).Debug(context.TODO(), "Hello world")
		log.(server.Logger).Debugf(context.TODO(), "Hello %s", "world")
	})

	t.Run("Error", func(t *testing.T) {
		log.(server.Logger).Print(context.TODO(), errors.New("Hello world"))
		log.(server.Logger).Printf(context.TODO(), "Hello %s", errors.New("world"))
	})

	t.Run("Args", func(t *testing.T) {
		log.(server.Logger).With("str", "b", "bool", false).Print(context.TODO(), errors.New("Hello world"))
	})
}
