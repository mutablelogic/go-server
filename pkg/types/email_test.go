package types_test

import (
	"fmt"
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_IsEmail(t *testing.T) {
	assert := assert.New(t)

	t.Run("Valid plain address", func(t *testing.T) {
		assert.True(types.IsEmail("user@example.com", nil, nil))
	})

	t.Run("Valid address with name", func(t *testing.T) {
		assert.True(types.IsEmail("John Doe <john@example.com>", nil, nil))
	})

	t.Run("Extract name and address", func(t *testing.T) {
		var name, email string
		ok := types.IsEmail("John Doe <john@example.com>", &name, &email)
		assert.True(ok)
		assert.Equal("John Doe", name)
		assert.Equal("john@example.com", email)
	})

	t.Run("Extract address only", func(t *testing.T) {
		var email string
		ok := types.IsEmail("user@example.com", nil, &email)
		assert.True(ok)
		assert.Equal("user@example.com", email)
	})

	t.Run("Extract name only", func(t *testing.T) {
		var name string
		ok := types.IsEmail("Jane Smith <jane@test.org>", &name, nil)
		assert.True(ok)
		assert.Equal("Jane Smith", name)
	})

	t.Run("Plain address has empty name", func(t *testing.T) {
		var name, email string
		ok := types.IsEmail("user@example.com", &name, &email)
		assert.True(ok)
		assert.Equal("", name)
		assert.Equal("user@example.com", email)
	})

	invalidTests := []struct {
		Input string
	}{
		{""},
		{"not-an-email"},
		{"@missing-local.com"},
		{"missing-domain@"},
		{"spaces in@address.com"},
	}

	for i, test := range invalidTests {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			assert.False(types.IsEmail(test.Input, nil, nil))
		})
	}
}
