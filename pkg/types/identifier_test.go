package types_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_Identifier_000(t *testing.T) {
	tests := []struct {
		In  string
		Out bool
	}{
		{"", false},
		{"a", true},
		{"0", false},
		{"a0", true},
		{"_", false},
		{"a_", true},
		{"a.b", false},
		{"a-b", true},
		{"-b", false},
	}

	for i, test := range tests {
		if out := types.IsIdentifier(test.In); out != test.Out {
			t.Errorf("Test %d: Expected %v, got %v for %q", i, test.Out, out, test.In)
		}
	}
}
