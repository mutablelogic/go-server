package types

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SchemeSecure             = "https"
	SchemeInsecure           = "http"
	ContentAcceptHeader      = "Accept"
	ContentTypeHeader        = "Content-Type"
	ContentLengthHeader      = "Content-Length"
	ContentDispositonHeader  = "Content-Disposition"
	ContentModifiedHeader    = "Last-Modified"
	ContentHashHeader        = "ETag"
	ContentPathHeader        = "X-Path"
	ContentNameHeader        = "X-Name"
	ContentDescriptionHeader = "X-Description"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RequestContentType(r *http.Request) (string, error) {
	return ParseContentType(r.Header.Get(ContentTypeHeader))
}

func AcceptContentType(r *http.Request) (string, error) {
	if accept := r.Header.Get(ContentAcceptHeader); accept == "" {
		return "", nil
	} else {
		return ParseContentType(accept)
	}
}
