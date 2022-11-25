package types_test

import (
	"context"
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
		{"test ${ test }", "test test"},
	}

	ctx := context.Background()
	for i, test := range tests {
		if out, err := test.In.Eval(ctx, t.Name()); err != nil {
			t.Error(err)
		} else if out != test.Out {
			t.Errorf("Test %d: Expected %q, got %q for %q", i, test.Out, out, test.In)
		}
	}
}
