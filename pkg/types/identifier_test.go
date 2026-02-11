package types_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_IsIdentifier(t *testing.T) {
	assert := assert.New(t)

	validTests := []string{
		"a",
		"abc",
		"myIdentifier",
		"my-identifier",
		"my_identifier",
		"A",
		"ABC",
		"a123",
		"node-main",
		"httpserver",
		"Z-0_a",
	}

	for _, input := range validTests {
		t.Run(fmt.Sprintf("Valid/%s", input), func(t *testing.T) {
			assert.True(types.IsIdentifier(input), "expected %q to be valid", input)
		})
	}

	invalidTests := []struct {
		Name  string
		Input string
	}{
		{"Empty", ""},
		{"StartsWithDigit", "1abc"},
		{"StartsWithHyphen", "-abc"},
		{"StartsWithUnderscore", "_abc"},
		{"ContainsDot", "a.b"},
		{"ContainsSpace", "a b"},
		{"ContainsSlash", "a/b"},
		{"ContainsAt", "a@b"},
		{"TooLong", "a" + strings.Repeat("b", 64)},
	}

	for _, test := range invalidTests {
		t.Run(fmt.Sprintf("Invalid/%s", test.Name), func(t *testing.T) {
			assert.False(types.IsIdentifier(test.Input), "expected %q to be invalid", test.Input)
		})
	}

	t.Run("MaxLength64", func(t *testing.T) {
		assert.True(types.IsIdentifier("a" + strings.Repeat("b", 63)))
	})
}
