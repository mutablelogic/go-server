package types_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_label_000(t *testing.T) {
	tests := []struct {
		In  []string
		Out string
		Err bool
	}{
		{[]string{""}, "", true},
		{[]string{"a"}, "a", false},
		{[]string{"a", "b"}, "a.b", false},
		{[]string{"a", "", "b"}, "a..b", true},
	}

	for _, test := range tests {
		t.Run(test.Out, func(t *testing.T) {
			label := types.NewLabel(test.In[0], test.In[1:]...)
			if label == "" {
				if !test.Err {
					t.Errorf("Expected label, got error for %q", test.Out)
				}
				return
			} else if label != types.Label(test.Out) {
				t.Errorf("Expected %v, got %v for %q", test.Out, label, test.In)
			} else if _, err := types.ParseLabel(string(label)); err != nil {
				t.Error(err)
			} else if prefix := label.Prefix(); prefix != "a" {
				t.Errorf("Expected prefix %q, got %q", "a", prefix)
			}
		})
	}
}
