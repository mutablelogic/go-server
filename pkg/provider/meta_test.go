package provider_test

import (
	"testing"

	// Packages
	provider "github.com/mutablelogic/go-server/pkg/provider"
	assert "github.com/stretchr/testify/assert"
)

type Test1 struct {
	A string
}

func Test_Meta_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		_, err := provider.NewMeta(nil, "")
		assert.Error(err)
	})
	t.Run("2", func(t *testing.T) {
		_, err := provider.NewMeta("hello, world", "")
		assert.Error(err)
	})

	t.Run("3", func(t *testing.T) {
		_, err := provider.NewMeta(Test1{}, "")
		assert.NoError(err)
	})

	t.Run("4", func(t *testing.T) {
		_, err := provider.NewMeta(&Test1{}, "")
		assert.NoError(err)
	})

}

func Test_Meta_002(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		meta, err := provider.NewMeta(Test1{}, "")
		assert.NoError(err)
		assert.NotNil(meta)
	})
}
