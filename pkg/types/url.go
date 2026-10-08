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
		return u.Opaque
	}

	p := u.Path
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && isASCIILetter(p[1]) {
		p = p[1:]
	}

	if host := u.Host; host != "" && !strings.EqualFold(host, "localhost") {
		return "//" + host + p
	}

	return p
}

// unicode.IsLetter would misread a UTF-8 continuation byte as a Latin-1 letter.
func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
