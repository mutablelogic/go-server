package httprequest

import (
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// FormValueMaxSize is the largest value of a form, other than a file,
	// which Binary reads
	FormValueMaxSize = 1 << 20 // 1 MB
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Binary reads the files uploaded with a request, as httpresponse.Binary
// writes a file in a response. It calls fn for each file, in order, with
// the name of its form field, the form's values which were before it, and
// the file, whose body is streamed from the request, so it's only read in
// fn, and isn't held in memory or in a temporary file, as Read holds it.
//
// The parts of a multipart/form-data request with a file name are files,
// and its other parts are values, which are passed to fn as far as they've
// been read, so a value sent before a file, as browsers and go-client send
// a form's fields in order, is known with it. A value larger than
// FormValueMaxSize is an error. The body of any other request is one file,
// whose field name and path are empty, with the request's headers.
//
// A file's path keeps its directories, from its X-Path part header, as
// go-client sends it, or the file name in its Content-Disposition, as a
// browser sends a folder's files, cleaned, and without a leading slash, so it
// can be joined to a directory. A path which would be above it is the file
// name alone. A file's modified time is from its Last-Modified header, when
// it's set.
//
// A large file can take longer to read than the server's read timeout, so a
// caller can clear the read deadline before, with http.ResponseController.
// Returns the form's values, which are all of them when there's no error,
// and the error from fn, including an error reading a file's body, or
// reading the form's parts or values, which is a bad request.
func Binary(r *http.Request, fn func(field string, values url.Values, file types.File) error) (url.Values, error) {
	values := url.Values{}

	// A request which isn't a form is one file
	if contentType, _ := types.ParseContentType(r.Header.Get(types.ContentTypeHeader)); contentType != types.ContentTypeFormData {
		header := textproto.MIMEHeader(r.Header)
		return values, fn("", values, types.File{
			Body:        io.NopCloser(r.Body),
			ContentType: r.Header.Get(types.ContentTypeHeader),
			Header:      header,
			ModTime:     modTime(header),
		})
	}

	// Read the parts of the form, a part at a time
	reader, err := r.MultipartReader()
	if err != nil {
		return values, httpresponse.ErrBadRequest.Withf("%v", err)
	}
	for {
		// The form ends with its final boundary, which is io.EOF itself,
		// and a form which ends without it, or never starts, is an error
		// which wraps io.EOF
		part, err := reader.NextPart()
		if err == io.EOF {
			return values, nil
		} else if err != nil {
			return values, httpresponse.ErrBadRequest.Withf("%v", err)
		}

		// A part without a file name is a value. A part isn't closed on an
		// error, as closing it reads the rest of it.
		field, filename := part.FormName(), dispositionFilename(part.Header)
		if filename == "" {
			value, err := io.ReadAll(io.LimitReader(part, FormValueMaxSize+1))
			if err != nil {
				return values, httpresponse.ErrBadRequest.Withf("%v", err)
			} else if len(value) > FormValueMaxSize {
				return values, httpresponse.ErrBadRequest.Withf("value of %q is too large", field)
			}
			values.Add(field, string(value))
			continue
		}

		// A part with a file name is a file. When fn returns an error, the
		// rest of the request isn't read, so a file which is rejected,
		// which can be large, isn't read to its end.
		if err := fn(field, values, types.File{
			Path:        filePath(part.Header, filename),
			Body:        part,
			ContentType: part.Header.Get(types.ContentTypeHeader),
			Header:      part.Header,
			ModTime:     modTime(part.Header),
		}); err != nil {
			return values, err
		}
		part.Close()
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// dispositionFilename returns the file name of a part, from its
// Content-Disposition, with its directories, which multipart.Part.FileName
// removes, or an empty string when it has none
func dispositionFilename(header textproto.MIMEHeader) string {
	_, params, err := mime.ParseMediaType(header.Get("Content-Disposition"))
	if err != nil {
		return ""
	}
	return params["filename"]
}

// filePath returns the path of a file part, from its X-Path header, or else
// its file name, with a backslash, as some browsers send for Windows, as a
// separator, cleaned, and without a leading slash. A path which is empty,
// would be above its root, or isn't local on this system (see
// filepath.IsLocal), such as one with a drive letter or a reserved name on
// Windows, isn't used, and when neither is, it's the file name without its
// directories, when that's local, or else "file".
func filePath(header textproto.MIMEHeader, filename string) string {
	filename = strings.ReplaceAll(filename, `\`, "/")
	for _, p := range []string{header.Get(types.ContentPathHeader), filename} {
		if p = strings.TrimPrefix(path.Clean(p), "/"); p != "" && isLocal(p) {
			return p
		}
	}
	if name := path.Base(filename); isLocal(name) {
		return name
	}
	return "file"
}

// isLocal returns true when the path, separated by "/", is a path within a
// directory on this system, which isn't empty, absolute, or above it, and
// on Windows, doesn't have a drive letter, a colon, or a reserved name
func isLocal(p string) bool {
	return p != "" && p != "." && filepath.IsLocal(filepath.FromSlash(p))
}

// modTime returns the time of a Last-Modified header, or the zero time
func modTime(header textproto.MIMEHeader) (t time.Time) {
	if value := header.Get(types.ContentModifiedHeader); value != "" {
		t, _ = http.ParseTime(value)
	}
	return t
}
