package meta_test

import (
	"testing"

	// Packages

	"github.com/mutablelogic/go-server/pkg/parser/meta"
	assert "github.com/stretchr/testify/assert"
)

type Test1 struct {
	A string
}

func Test_Meta_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		_, err := meta.New(nil, "")
		assert.Error(err)
	})
	t.Run("2", func(t *testing.T) {
		_, err := meta.New("hello, world", "")
		assert.Error(err)
	})

	t.Run("3", func(t *testing.T) {
		_, err := meta.New(Test1{}, "")
		assert.NoError(err)
	})

	t.Run("4", func(t *testing.T) {
		_, err := meta.New(&Test1{}, "")
		assert.NoError(err)
	})

}

func Test_Meta_002(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		meta, err := meta.New(Test1{}, "")
		assert.NoError(err)
		assert.NotNil(meta)
	})
}
