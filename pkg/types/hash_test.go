package types_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Hash(t *testing.T) {
	assert := assert.New(t)

	t.Run("Empty input", func(t *testing.T) {
		result := types.Hash([]byte{})
		assert.Equal("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", result)
	})

	t.Run("Known value", func(t *testing.T) {
		result := types.Hash([]byte("hello"))
		assert.Equal("2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", result)
	})

	t.Run("Deterministic", func(t *testing.T) {
		data := []byte("test data for hashing")
		assert.Equal(types.Hash(data), types.Hash(data))
	})

	t.Run("Different inputs differ", func(t *testing.T) {
		assert.NotEqual(types.Hash([]byte("a")), types.Hash([]byte("b")))
	})

	t.Run("Returns hex string of length 64", func(t *testing.T) {
		result := types.Hash([]byte("anything"))
		assert.Len(result, 64)
	})
}
