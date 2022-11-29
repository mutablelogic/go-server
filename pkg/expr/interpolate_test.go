package expr_test

import (
	"fmt"
	"strings"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-server/pkg/expr"
)

func Test_Interpolate_001(t *testing.T) {
	// Non-error cases
	tests := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"}", `str<"}">`},
		{"test}", `str<"test}">`},
		{"test}test", `str<"test}test">`},
		{"{test}{test}", `str<"{test}{test}">`},
		{"$${", `str<"${">`},
		{"${}", ""},
		{"${test}", `expr<"test">`},
		{"}test}test}", `str<"}test}test}">`},
		{"}test}$${test}", `str<"}test}${test}">`},
		{"\n", `str<"\n">`},
		{"${test1}$", `expr<"test1">str<"$">`},
		{"${test1}${test2}", `expr<"test1">expr<"test2">`},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			assert := assert.New(t)
			tokens, err := Interpolate(strings.NewReader(test.in), Pos{Path: nil, Line: 1, Col: 1})
			assert.NoError(err)
			assert.Equal(test.out, toString(tokens))
		})
	}
}

func toString(v []*Token) string {
	var str string
	for _, token := range v {
		str += fmt.Sprint(token)
	}
	return str
}
