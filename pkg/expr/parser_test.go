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

func Test_Parser_001(t *testing.T) {
	// Non-error cases
	tests := []struct {
		in, out string
	}{
		{"0 1 2 3 4 5 6 7 8 9", "0,1,2,3,4,5,6,7,8,9"},
		{"0.1 1.2 3. 4. 567.345", "0.1,1.2,3,4,567.345"},
		{"+1 +1.2 +233. +34", "1,1.2,233,34"},
		{"-1 -1.2 -233. -34", "-1,-1.2,-233,-34"},
		{"!1 !2 !2.3", "Not(1),Not(2),Not(2.3)"},
		{"5 + 5", "Plus(5,5)"},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			assert := assert.New(t)
			parser := NewParser(strings.NewReader(test.in), Pos{})
			assert.NotNil(parser)
			node, err := parser.Parse()
			assert.NoError(err)
			assert.Equal(test.out, fmt.Sprint(node))
		})
	}
}
