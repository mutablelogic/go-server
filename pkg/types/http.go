package types

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SchemeSecure            = "https"
	SchemeInsecure          = "http"
	ContentTypeHeader       = "Content-Type"
	ContentAcceptHeader     = "Accept"
	ContentLengthHeader     = "Content-Length"
	ContentDispositonHeader = "Content-Disposition"
	ContentModifiedHeader   = "Last-Modified"
	ContentHashHeader       = "ETag"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RequestContentType(r *http.Request) (string, error) {
	return ParseContentType(r.Header.Get(ContentTypeHeader))
}

func AcceptContentType(r *http.Request) (string, error) {
	return ParseContentType(r.Header.Get(ContentAcceptHeader))
}
