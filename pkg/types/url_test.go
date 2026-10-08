package types_test

import (
	"net/url"
	"testing"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

func TestFilePath(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"posix, RFC 8089 triple-slash", "file:///tmp/sample.mp3", "/tmp/sample.mp3"},
		{"posix, single-slash opaque-looking form", "file:/tmp/sample.mp3", "/tmp/sample.mp3"},
		{"windows, RFC 8089 triple-slash", "file:///D:/a/go-media/sample.mp3", "D:/a/go-media/sample.mp3"},
		{"windows, opaque form (no //)", "file:D:/a/go-media/sample.mp3", "D:/a/go-media/sample.mp3"},
		{"ordinary posix path isn't mistaken for a drive letter", "file:///foo/bar", "/foo/bar"},
		{"unc path preserves the host", "file://server/share/media.mp3", "//server/share/media.mp3"},
		{"explicit localhost authority is treated as local", "file://localhost/tmp/sample.mp3", "/tmp/sample.mp3"},
		{"localhost is case-insensitive", "file://LOCALHOST/tmp/sample.mp3", "/tmp/sample.mp3"},
		{"wrong scheme returns empty", "https://example.com/sample.mp3", ""},
		{"device scheme returns empty", "device://avfoundation/0:0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("url.Parse(%q): %v", tc.url, err)
			}
			if got := types.FilePath(u); got != tc.want {
				t.Fatalf("FilePath(%q) = %q, want %q", tc.url, got, tc.want)
			}
		})
	}
}

func TestFilePath_Nil(t *testing.T) {
	if got := types.FilePath(nil); got != "" {
		t.Fatalf("FilePath(nil) = %q, want \"\"", got)
	}
}
