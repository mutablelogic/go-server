package types

import (
	"net/url"
	"strings"
)

// FilePath returns a Windows-compliant filesystem path from a parsed URL
// with no scheme or an explicit "file" scheme; "" for a nil URL or any
// other scheme.
func FilePath(u *url.URL) string {
	if u == nil || (u.Scheme != "" && u.Scheme != "file") {
		return ""
	}

	if u.Opaque != "" {
		// Opaque is the raw, still percent-encoded text after the scheme
		// colon (e.g. "C:/Program%20Files/a.mp3") - unlike Path, net/url
		// never decodes it, so it must be unescaped here to land on the
		// same OS path the hierarchical form (file:///...) would produce.
		p, err := url.PathUnescape(u.Opaque)
		if err != nil {
			return ""
		}
		return p
	}

	p := u.Path

	// A UNC path never carries a drive letter, so the stripping below must
	// not run on one - applying it first would corrupt a share path that
	// happens to start with a drive-letter-like segment, e.g.
	// file://server/C:/share (Path="/C:/share") becoming "//serverC:/share"
	// instead of "//server/C:/share".
	if host := u.Host; host != "" && !strings.EqualFold(host, "localhost") {
		return "//" + host + p
	}

	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && isASCIILetter(p[1]) {
		p = p[1:]
	}

	return p
}

// unicode.IsLetter would misread a UTF-8 continuation byte as a Latin-1 letter.
func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
