package meta_test

import (
	"context"
	"testing"

	// Packages
	server "github.com/mutablelogic/go-server"
	meta "github.com/mutablelogic/go-server/pkg/provider/meta"
	assert "github.com/stretchr/testify/assert"
)

type Test1 struct {
	A string
}

func (t Test1) Name() string {
	return "Test1"
}
func (t Test1) Description() string {
	return "Test1"
}
func (t Test1) New(context.Context) (server.Task, error) {
	return nil, nil
}

var _ server.Plugin = Test1{}

func Test_Meta_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		_, err := meta.New(nil)
		assert.Error(err)
	})

	t.Run("2", func(t *testing.T) {
		_, err := meta.New(Test1{})
		assert.NoError(err)
	})

	t.Run("3", func(t *testing.T) {
		_, err := meta.New(&Test1{})
		assert.NoError(err)
	})

}

func Test_Meta_002(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		meta, err := meta.New(Test1{})
		assert.NoError(err)
		assert.NotNil(meta)
	})
}
