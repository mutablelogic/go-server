package version

import (
	"runtime/debug"
	"testing"
)

func Test_moduleVersion(t *testing.T) {
	for version, want := range map[string]string{
		"":        "",
		"(devel)": "",
		"v1.0.0":  "v1.0.0",
		"v0.0.0-20261118123708-ead0a2da1cc5+dirty": "v0.0.0-20261118123708-ead0a2da1cc5+dirty",
	} {
		info := &debug.BuildInfo{Main: debug.Module{Path: "example.com/cmd", Version: version}}
		if got := moduleVersion(info); got != want {
			t.Errorf("moduleVersion(%q) = %q, want %q", version, got, want)
		}
	}
	if got := moduleVersion(nil); got != "" {
		t.Errorf("moduleVersion(nil) = %q, want none", got)
	}
}
