package httprequest

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
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
	// which ReadFiles reads
	FormValueMaxSize = 1 << 20 // 1 MB
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ReadFiles calls fn for each file uploaded with the request, in order, with
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
func ReadFiles(r *http.Request, fn func(field string, values url.Values, file types.File) error) (url.Values, error) {
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
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			return values, nil
		} else if err != nil {
			return values, httpresponse.ErrBadRequest.Withf("%v", err)
		}

		// A part without a file name is a value
		field, filename := part.FormName(), dispositionFilename(part.Header)
		if filename == "" {
			value, err := io.ReadAll(io.LimitReader(part, FormValueMaxSize+1))
			part.Close()
			if err != nil {
				return values, httpresponse.ErrBadRequest.Withf("%v", err)
			} else if len(value) > FormValueMaxSize {
				return values, httpresponse.ErrBadRequest.Withf("value of %q is too large", field)
			}
			values.Add(field, string(value))
			continue
		}

		// A part with a file name is a file
		err = fn(field, values, types.File{
			Path:        filePath(part.Header, filename),
			Body:        part,
			ContentType: part.Header.Get(types.ContentTypeHeader),
			Header:      part.Header,
			ModTime:     modTime(part.Header),
		})
		part.Close()
		if err != nil {
			return values, err
		}
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
// or would be above its root, isn't used, and when neither is, it's the file
// name without its directories, or "file".
func filePath(header textproto.MIMEHeader, filename string) string {
	filename = strings.ReplaceAll(filename, `\`, "/")
	for _, p := range []string{header.Get(types.ContentPathHeader), filename} {
		if p = strings.TrimPrefix(path.Clean(p), "/"); p != "" && p != "." && p != ".." && !strings.HasPrefix(p, "../") {
			return p
		}
	}
	if name := path.Base(filename); name != "" && name != "." && name != ".." && name != "/" {
		return name
	}
	return "file"
}

// modTime returns the time of a Last-Modified header, or the zero time
func modTime(header textproto.MIMEHeader) (t time.Time) {
	if value := header.Get(types.ContentModifiedHeader); value != "" {
		t, _ = http.ParseTime(value)
	}
	return t
}
