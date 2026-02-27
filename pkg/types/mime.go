package types

import (
	"mime"
	"regexp"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeJSON       = "application/json"
	ContentTypeXML        = "application/xml"
	ContentTypeRSS        = "application/rss+xml"
	ContentTypeBinary     = "application/octet-stream"
	ContentTypeCSV        = "text/csv"
	ContentTypeTextXml    = "text/xml"
	ContentTypeTextPlain  = "text/plain"
	ContentTypeTextStream = "text/event-stream"
	ContentTypeFormData   = "multipart/form-data"
	ContentTypeAny        = "*/*"
)

// reHeaderKey matches a valid RFC 7230 header field name token:
// one or more "tchar" characters as defined in RFC 7230 §3.2.6.
var reHeaderKey = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Parse a content type header to return the mimetype
func ParseContentType(header string) (string, error) {
	mimetype, _, err := mime.ParseMediaType(header)
	return mimetype, err
}

// IsValidHeaderKey reports whether s is a valid RFC 7230 header field name
// token: one or more tchar characters.
func IsValidHeaderKey(s string) bool {
	return reHeaderKey.MatchString(s)
}
