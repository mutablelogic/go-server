package types_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_dns_001(t *testing.T) {
	tests := []struct {
		In  []string
		Out string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"local"}, "local."},
		{[]string{"."}, "."},
		{[]string{".."}, "."},
		{[]string{".", "."}, ".."},
		{[]string{"host", "name"}, "host.name."},
		{[]string{"host"}, "host."},
		{[]string{".host."}, "host."},
		{[]string{".host.", "name", "local"}, "host.name.local."},
	}
	for _, test := range tests {
		if out := types.Fqn(test.In...); out != test.Out {
			t.Errorf("Fqn(%q) returned %q, expected %q", test.In, out, test.Out)
		}
	}
}

func Test_dns_002(t *testing.T) {
	tests := []struct {
		In, Domain, Out string
	}{
		{"", "", ""},
		{"dns-sd.local", "local", "dns-sd"},
		{"dns-sd.local.", ".local.", "dns-sd"},
		{"dns-sd.local.", "local.", "dns-sd"},
		{"dns-sd.test", "local.", "dns-sd.test"},
		{"dns-sd.test.", "local.", "dns-sd.test"},
		{"dns-sd.test.local", "local.", "dns-sd.test"},
		{"..dns-sd.test.local", "local.", "dns-sd.test"},
		{"..dns-sd.test.local..", "", "dns-sd.test.local"},
	}
	for _, test := range tests {
		if out := types.Unfqn(test.In, test.Domain); out != test.Out {
			t.Errorf("Unfqn(%q,%q) returned %q, expected %q", test.In, test.Domain, out, test.Out)
		}
	}
}
