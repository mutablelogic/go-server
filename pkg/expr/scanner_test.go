package expr_test

import (
	"strings"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-server/pkg/expr"
)

func Test_Scanner_001(t *testing.T) {
	// Non-error cases
	tests := []struct {
		in string
	}{
		{""},
		{"0 1 2 3 4 5 6 7 8 9"},
		{"func(test)"},
		{"   var.test   "},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			assert := assert.New(t)
			scanner := NewScanner(strings.NewReader(test.in), Pos{})
			assert.NotNil(scanner)
			for tok := scanner.Scan(); tok.Kind != EOF; tok = scanner.Scan() {
				assert.NotNil(tok)
			}
		})
	}
}

func Test_Scanner_002(t *testing.T) {
	// Non-error cases - general
	tests := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"0 1", "Number Space Number"},
		{"func(test)", "Ident OpenParen Ident CloseParen"},
		{"   var.test   ", "Space Ident Punkt Ident Space"},
		{`'test'`, "String"},
		{`'te"st'`, "String"},
		{`"e'st"`, "String"},
		{`"e\"st"`, "String"},
		{`'e\'st'`, "String"},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			assert := assert.New(t)
			tokens, err := NewScanner(strings.NewReader(test.in), Pos{}).Tokens()
			assert.NoError(err)
			assert.NotNil(tokens)
			assert.Equal(test.out, toTokenString(tokens))
		})
	}
}

func toTokenString(tokens []*Token) string {
	var result []string
	for _, token := range tokens {
		result = append(result, token.Kind.String())
	}
	return strings.Join(result, " ")
}
