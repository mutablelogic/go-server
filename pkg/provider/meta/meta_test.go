package meta_test

import (
	"context"
	"os"
	"testing"

	// Packages
	server "github.com/mutablelogic/go-server"
	meta "github.com/mutablelogic/go-server/pkg/provider/meta"
	assert "github.com/stretchr/testify/assert"
)

type Test1 struct {
	A string
}

type Test2 struct {
	A string
	B struct {
		C string `name:"c"`
		D string `name:"d" default:"default"`
	} `json:"b" help:"this is a block"`
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

func (t Test2) Name() string {
	return "Test2"
}
func (t Test2) Description() string {
	return "Test2"
}
func (t Test2) New(context.Context) (server.Task, error) {
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

		meta.Write(os.Stderr)

		assert.Equal("Test1", meta.Name)
		assert.Equal("Test1", meta.Description)
		field := meta.Get("Test1.A")
		assert.NotNil(field)
		assert.Equal("A", field.Name)
		assert.Equal("Test1.A", field.Label())
		assert.Equal("", field.Description)
	})

	t.Run("2", func(t *testing.T) {
		meta, err := meta.New(Test2{})
		assert.NoError(err)
		assert.NotNil(meta)
		assert.Equal("Test2", meta.Name)
		assert.Equal("Test2", meta.Description)

		meta.Write(os.Stderr)

		field := meta.Get("Test2.A")
		assert.NotNil(field)
		assert.Equal("A", field.Name)
		assert.Equal("Test2.A", field.Label())
		assert.Equal("", field.Description)

		field = meta.Get("Test2.b")
		assert.NotNil(field)
		assert.Nil(field.Type)
		assert.Equal("b", field.Name)
		assert.Equal("Test2.b", field.Label())
		assert.Equal("this is a block", field.Description)

		field = meta.Get("Test2.b.c")
		assert.NotNil(field)
		assert.NotNil(field.Type)
		assert.Equal("c", field.Name)
		assert.Equal("Test2.b.c", field.Label())

		field = meta.Get("Test2.b.d")
		assert.NotNil(field)
		assert.NotNil(field.Type)
		assert.Equal("d", field.Name)
		assert.Equal("Test2.b.d", field.Label())
	})
}
