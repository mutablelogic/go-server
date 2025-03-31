package types

import (
	"mime"
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

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Parse a content type header to return the mimetype
func ParseContentType(header string) (string, error) {
	mimetype, _, err := mime.ParseMediaType(header)
	return mimetype, err
}
