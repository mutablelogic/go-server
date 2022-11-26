package expr_test

import (
	"strings"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-server/pkg/expr"
)

func Test_Lexer_001(t *testing.T) {
	// Non-error cases
	tests := []struct {
		in string
	}{
		{""},
		{"0 1 2 3 4 5 6 7 8 9"},
		{"   var.test   "},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			assert := assert.New(t)
			_, err := Lexer(strings.NewReader(test.in), Pos{Path: nil, Line: 1, Col: 1})
			assert.NoError(err)
		})
	}
}
