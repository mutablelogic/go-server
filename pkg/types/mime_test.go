package types_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Mime_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		_, err := types.ParseContentType("")
		if assert.Error(err) {
			assert.ErrorContains(err, "no media type")
		}
	})

	t.Run("2", func(t *testing.T) {
		mime, err := types.ParseContentType(types.ContentTypeJSON)
		if assert.NoError(err) {
			assert.Equal(mime, types.ContentTypeJSON)
		}
	})
}

func Test_IsValidHeaderKey(t *testing.T) {
	assert := assert.New(t)

	valid := []string{
		"Content-Type",
		"X-Custom-Header",
		"Accept",
		"X-Meta-Foo",
		"a",
		"a1",
		"x-foo.bar",
		"~token~",
	}
	for _, k := range valid {
		assert.True(types.IsValidHeaderKey(k), "expected valid: %q", k)
	}

	invalid := []string{
		"",
		"has space",
		"has\ttab",
		"has\nnewline",
		"has(paren)",
		"has/slash",
		"has:colon",
		"has@at",
		"has[bracket]",
		"has{brace}",
		"has\"quote",
		"has\\backslash",
		"\x00null",
		"\x80non-ascii",
	}
	for _, k := range invalid {
		assert.False(types.IsValidHeaderKey(k), "expected invalid: %q", k)
	}
}
