package httprequest

import (
	"encoding/json"
	"mime"
	"net/http"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeJson           = "application/json"
	ContentTypeTextXml        = "text/xml"
	ContentTypeApplicationXml = "application/xml"
	ContentTypeText           = "text/"
	ContentTypeBinary         = "application/octet-stream"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Read(r *http.Request, v interface{}) error {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	switch r.Header.Get("Content-Type") {
	case ContentTypeJson:
		return readJson(r, v)
	}
	return ErrBadParameter.Withf("unexpected content type %q", contentType)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readJson(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
