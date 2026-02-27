package types

import (
	"io"
	"net/textproto"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// File represents a file part received from a multipart/form-data request, or
// to be sent as part of a multipart/form-data request. ContentType is optional;
// when set it is used as the part Content-Type instead of the default
// application/octet-stream. Header holds all part-level MIME headers
// (e.g. Content-Disposition, Content-Type, and any custom headers); it is
// populated when decoding and merged during encoding.
type File struct {
	Path        string
	Body        io.Reader
	ContentType string               // optional MIME type
	Header      textproto.MIMEHeader // all part-level MIME headers
}
