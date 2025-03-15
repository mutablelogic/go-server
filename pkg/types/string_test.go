package types_test

import (
	"fmt"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_IsSingleQuoted_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		assert.False(types.IsSingleQuoted(""))
	})

	t.Run("2", func(t *testing.T) {
		assert.False(types.IsSingleQuoted("'"))
	})
	t.Run("3", func(t *testing.T) {
		assert.False(types.IsSingleQuoted("'ddd"))
	})
	t.Run("4", func(t *testing.T) {
		assert.True(types.IsSingleQuoted("'ddd'"))
	})

	t.Run("5", func(t *testing.T) {
		assert.False(types.IsSingleQuoted("ddd'"))
	})
}

func Test_IsDoubleQuoted_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		assert.False(types.IsDoubleQuoted(""))
	})

	t.Run("2", func(t *testing.T) {
		assert.False(types.IsDoubleQuoted("\""))
	})
	t.Run("3", func(t *testing.T) {
		assert.False(types.IsDoubleQuoted("\"ddd"))
	})
	t.Run("4", func(t *testing.T) {
		assert.True(types.IsDoubleQuoted("\"ddd\""))
	})

	t.Run("5", func(t *testing.T) {
		assert.False(types.IsDoubleQuoted("ddd\""))
	})
}

func Test_Quote(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		Expected string
		Input    string
	}{
		{"''", ""},
		{"'abc'", "abc"},
		{"''''", "'"},
		{"'ab''cd'", "ab'cd"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.Equal(test.Expected, types.Quote(test.Input))
		})
	}
}

func Test_DoubleQuote(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		Expected string
		Input    string
	}{
		{`""`, ``},
		{`"abc"`, `abc`},
		{`""""`, `"`},
		{`"ab""cd"`, `ab"cd`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.Equal(test.Expected, types.DoubleQuote(test.Input))
		})
	}
}

func Test_IsNumeric(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		Expected bool
		Input    string
	}{
		{false, ""},
		{false, "100.0"},
		{false, "100X"},
		{true, "0"},
		{true, "100"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.Equal(test.Expected, types.IsNumeric(test.Input))
		})
	}
}

func Test_IsUppercase(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		Expected bool
		Input    string
	}{
		{false, ""},
		{false, "ABCd"},
		{false, "123"},
		{true, "ABC"},
		{true, "D"},
		{false, "B123"},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			assert.Equal(test.Expected, types.IsUppercase(test.Input))
		})
	}
}
