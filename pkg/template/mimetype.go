package template

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ContentTypeDetect struct {
	sync.Mutex
	buf []byte
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultBufSize     = 512
	defaultContentType = "application/octet-stream"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewContentTypeDetect() *ContentTypeDetect {
	this := new(ContentTypeDetect)
	this.buf = make([]byte, defaultBufSize)
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// DetectContentType returns the detected content type and optionally character set of
// the given reader, assuming the reader is at the beginning of the file. Will
// return an empty string if the file has no content, or application/octet-stream
// if the filetype could not be determined.
func (m *ContentTypeDetect) DetectContentType(r io.Reader, info fs.FileInfo) (string, string, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	n, err := r.Read(m.buf)
	if n == 0 {
		return "", "", nil
	}
	if err != nil {
		return defaultContentType, "", err
	}
	if mimetype := http.DetectContentType(m.buf); mimetype != defaultContentType {
		if mediatype, params, err := mime.ParseMediaType(mimetype); err == nil {
			return mediatype, params["charset"], nil
		}
	}
	if info != nil {
		if mimetype := mime.TypeByExtension(filepath.Ext(info.Name())); mimetype != "" {
			if mediatype, params, err := mime.ParseMediaType(mimetype); err == nil {
				return mediatype, params["charset"], nil
			}
		}
	}
	return defaultContentType, "", nil
}
