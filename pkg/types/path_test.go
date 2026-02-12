package types_test

import (
	"fmt"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
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

func Test_JoinPath(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		Prefix   string
		Path     string
		Expected string
	}{
		{"", "", "/"},
		{"/", "", "/"},
		{"", "/", "/"},
		{"/api", "/v1", "/api/v1"},
		{"/api", "v1", "/api/v1"},
		{"/api/", "/v1", "/api/v1"},
		{"/api/", "v1/", "/api/v1"},
		{"api", "v1", "/api/v1"},
		{"/", "/path", "/path"},
		{"/prefix", "/", "/prefix"},
		{"/a/b", "/c/d", "/a/b/c/d"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert.Equal(test.Expected, types.JoinPath(test.Prefix, test.Path))
		})
	}
}
