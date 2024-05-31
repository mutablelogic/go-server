package types_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_Filename_000(t *testing.T) {
	tests := []struct {
		In  string
		Out bool
	}{
		{"", false},
		{"ab", true},
		{"00", true},
		{"a0", true},
		{"_a", false},
		{"a_", false},
		{"a.b", true},
		{"a-b", true},
		{"-b", false},
		{"b.conf", true},
		{"b.conf.", false},
	}

	for i, test := range tests {
		t.Run(test.In, func(t *testing.T) {
			if out := types.IsFilename(test.In); out != test.Out {
				t.Errorf("Test %d: Expected %v, got %v for %q", i, test.Out, out, test.In)
			}
		})
	}
}
