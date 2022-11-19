package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_Duration_000(t *testing.T) {
	tests := []struct {
		In  string
		Out types.Duration
	}{
		{"1s", types.Duration(time.Second)},
		{"1m", types.Duration(time.Minute)},
		{"1h", types.Duration(time.Hour)},
	}

	for i, test := range tests {
		var out types.Duration
		if err := json.Unmarshal([]byte(`"`+test.In+`"`), &out); err != nil {
			t.Errorf("Test %d: Expected no error, got %v for %q", i, err, test.In)
		} else if out != test.Out {
			t.Errorf("Test %d: Expected %v, got %v for %q", i, test.Out, out, test.In)
		}
	}
}
