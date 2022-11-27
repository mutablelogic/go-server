package types_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_Eval_000(t *testing.T) {
	tests := []struct {
		In  types.Eval
		Out string
	}{
		{"", ""},
		{"test", "test"},
		{"${}", ""},
		{"$${}", ""},
		{"test ${ test }", "test test"},
		{"test ${ test } test", "test test test"},
		{"test ${ test ", "test test"},
	}

	for i, test := range tests {
		if out, err := test.In.Interpolate(); err != nil {
			t.Error(err)
		} else {
			t.Logf("Test %d: %q => %v", i, test.In, out)
		}
	}
}
