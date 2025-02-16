package types_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-service/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Path_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		path := types.NormalisePath("")
		assert.Equal("/", path)
	})

	t.Run("2", func(t *testing.T) {
		path := types.NormalisePath("/")
		assert.Equal("/", path)
	})

	t.Run("3", func(t *testing.T) {
		path := types.NormalisePath("path")
		assert.Equal("/path", path)
	})

	t.Run("4", func(t *testing.T) {
		path := types.NormalisePath("path/")
		assert.Equal("/path", path)
	})

	t.Run("5", func(t *testing.T) {
		path := types.NormalisePath("/path/")
		assert.Equal("/path", path)
	})

	t.Run("6", func(t *testing.T) {
		path := types.NormalisePath("//path/")
		assert.Equal("/path", path)
	})

	t.Run("7", func(t *testing.T) {
		path := types.NormalisePath("path//")
		assert.Equal("/path", path)
	})
}
