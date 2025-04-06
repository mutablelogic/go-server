package logger_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	// Packages

	"github.com/mutablelogic/go-server/pkg/logger"
	assert "github.com/stretchr/testify/assert"
)

func Test_Logger(t *testing.T) {
	assert := assert.New(t)
	buf := strings.Builder{}
	log := logger.New(&buf, true)
	assert.NotNil(log)

	t.Run("Info", func(t *testing.T) {
		log.Print(context.TODO(), "Hello world")
		log.Printf(context.TODO(), "Hello %s", "world")
	})

	t.Run("Debug", func(t *testing.T) {
		log.Debug(context.TODO(), "Hello world")
		log.Debugf(context.TODO(), "Hello %s", "world")
	})

	t.Run("Error", func(t *testing.T) {
		log.Print(context.TODO(), errors.New("Hello world"))
		log.Printf(context.TODO(), "Hello %s", errors.New("world"))
	})

	t.Run("Args1", func(t *testing.T) {
		log.With("str", "b", "bool", false).Print(context.TODO(), errors.New("Hello world"))
	})

	t.Run("Args2", func(t *testing.T) {
		log.With("str", "b", "bool", false).With("int", 42).Print(context.TODO(), errors.New("Hello world"))
	})

	t.Log(buf.String())
}
