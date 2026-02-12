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

func Test_Unquote(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{"DoubleQuoted", `"hello"`, "hello"},
		{"WithEscape", `"hello\nworld"`, "hello\nworld"},
		{"BackQuoted", "`raw string`", "raw string"},
		{"NotQuoted", "plain", "plain"},
		{"Empty", "", ""},
		{"SingleChar", "x", "x"},
		{"EmptyQuoted", `""`, ""},
		{"WithTab", `"a\tb"`, "a\tb"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(test.Expected, types.Unquote(test.Input))
		})
	}
}

func Test_Stringify(t *testing.T) {
	assert := assert.New(t)

	t.Run("String", func(t *testing.T) {
		result := types.Stringify("hello")
		assert.Equal(`"hello"`, result)
	})

	t.Run("Int", func(t *testing.T) {
		result := types.Stringify(42)
		assert.Equal("42", result)
	})

	t.Run("Bool", func(t *testing.T) {
		result := types.Stringify(true)
		assert.Equal("true", result)
	})

	t.Run("Nil", func(t *testing.T) {
		result := types.Stringify[any](nil)
		assert.Equal("null", result)
	})

	t.Run("Map", func(t *testing.T) {
		result := types.Stringify(map[string]int{"a": 1})
		assert.Contains(result, `"a": 1`)
	})

	t.Run("Slice", func(t *testing.T) {
		result := types.Stringify([]int{1, 2, 3})
		assert.Contains(result, "1")
		assert.Contains(result, "2")
		assert.Contains(result, "3")
	})

	t.Run("Struct", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		result := types.Stringify(testStruct{Name: "Alice", Age: 30})
		assert.Contains(result, `"name": "Alice"`)
		assert.Contains(result, `"age": 30`)
	})
}
