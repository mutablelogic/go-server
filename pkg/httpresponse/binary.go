package httpresponse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Disposition is how a browser presents content: inline, in the browser, or
// as an attachment, which it saves
type Disposition string

// Binary describes content, such as a file, for its response
type Binary struct {
	// ContentType is the Content-Type, which is application/octet-stream when
	// it's empty
	ContentType string

	// Size is the Content-Length, when it's known
	Size *int64

	// ModTime is the Last-Modified time, when it isn't zero
	ModTime time.Time

	// ETag is the entity tag, without quotes, when it isn't empty
	ETag string

	// Path is the path of the content, whose file name is in the
	// Content-Disposition, when it isn't empty
	Path string

	// Disposition is how the browser presents the content, which is in the
	// Content-Disposition, with the file name, when it isn't empty
	Disposition Disposition

	// CacheControl is how the response is cached, which is private,
	// no-cache when it's the zero value
	CacheControl CacheControl

	// Meta is the metadata of the content, as JSON in the X-Meta header, when
	// it isn't nil
	Meta any
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DispositionInline     Disposition = "inline"
	DispositionAttachment Disposition = "attachment"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write responds with the content's headers, and for a GET request, its
// content, read from body. The request's preconditions are evaluated in the
// order of RFC 9110: when If-Match, or without it, If-Unmodified-Since,
// fails, it's 412 Precondition Failed, and when the content isn't modified
// (see NotModified), it's 304 Not Modified, with its headers. For a HEAD
// request it's its headers alone, and for a GET request without a body, 204
// No Content. Otherwise its content is written without the server's write
// timeout, as a large file, or a video a browser reads as it plays it, can
// take longer, in ranges when body can seek, such as a file.
//
// The content isn't sniffed as another type, and content which can run
// scripts in a browser, which is HTML and XML, including SVG, is sandboxed,
// so it can't run scripts with access to the server. Other types aren't, as
// a browser shows media and PDFs in a page of its own, which can't load
// them when it's sandboxed.
//
// Returns an error, without writing the response, when the headers can't be
// set, and otherwise any error reading or writing the content, after which
// the response can't be changed.
func (b Binary) Write(w http.ResponseWriter, r *http.Request, body io.Reader) error {
	// A precondition which fails has no content
	if !b.preconditions(r) {
		return Error(w, Err(http.StatusPreconditionFailed))
	}
	if err := b.setHeaders(w.Header()); err != nil {
		return err
	}

	// Not modified, the headers alone, or no content
	switch {
	case b.NotModified(r):
		w.WriteHeader(http.StatusNotModified)
		return nil
	case r.Method == http.MethodHead:
		w.WriteHeader(http.StatusOK)
		return nil
	case body == nil:
		w.Header().Del(types.ContentLengthHeader)
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// The content can take longer to write than the server's write timeout
	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}

	// Content which can seek is served in ranges, which also responds to
	// If-Range, and otherwise the whole content is written. The request's
	// preconditions have been evaluated, so ServeContent, which evaluates
	// them again, with its own parsing of etags, doesn't have them. It
	// doesn't return the errors reading or writing the content, so they're
	// recorded.
	if seeker, ok := body.(io.ReadSeeker); ok {
		req := *r
		req.Header = r.Header.Clone()
		for _, key := range []string{"If-Match", "If-None-Match", "If-Modified-Since", "If-Unmodified-Since"} {
			req.Header.Del(key)
		}
		reader, writer := &readSeekRecorder{ReadSeeker: seeker}, &writeRecorder{ResponseWriter: w}
		http.ServeContent(writer, &req, "", b.ModTime, reader)
		return errors.Join(reader.err, writer.err)
	} else {
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, body)
		return err
	}
}

// NotModified returns true when a GET or HEAD request's If-None-Match has
// the content's etag, by weak comparison, or is *, or without If-None-Match,
// the content hasn't been modified since If-Modified-Since, to the second,
// as HTTP dates are. An If-None-Match which can't be parsed doesn't match.
func (b Binary) NotModified(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	if match := r.Header.Get("If-None-Match"); match != "" {
		return b.matches(match, false)
	}
	if b.ModTime.IsZero() {
		return false
	}
	if since, err := http.ParseTime(r.Header.Get("If-Modified-Since")); err == nil {
		return !b.ModTime.Truncate(time.Second).After(since)
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// setHeaders sets the headers of the content
func (b Binary) setHeaders(header http.Header) error {
	if b.Meta != nil {
		meta, err := headerJSON(b.Meta)
		if err != nil {
			return err
		}
		header.Set(types.ContentMetaHeader, meta)
	}
	contentType := b.ContentType
	if contentType == "" {
		contentType = types.ContentTypeBinary
	}
	header.Set(types.ContentTypeHeader, contentType)
	if b.Size != nil {
		header.Set(types.ContentLengthHeader, strconv.FormatInt(*b.Size, 10))
	}
	if !b.ModTime.IsZero() {
		header.Set(types.ContentModifiedHeader, b.ModTime.UTC().Format(http.TimeFormat))
	}
	if b.ETag != "" {
		header.Set(types.ContentHashHeader, `"`+b.ETag+`"`)
	}
	header.Set("Cache-Control", b.CacheControl.String())
	if b.Disposition != "" {
		params := map[string]string{}
		if name := path.Base(b.Path); b.Path != "" && name != "." && name != "/" {
			params["filename"] = name
		}
		if disposition := mime.FormatMediaType(string(b.Disposition), params); disposition != "" {
			header.Set(types.ContentDispositonHeader, disposition)
		}
	}

	// The content isn't sniffed, and when it can run scripts, it's sandboxed
	header.Set("X-Content-Type-Options", "nosniff")
	if scriptable(contentType) {
		header.Set("Content-Security-Policy", "sandbox")
	}
	return nil
}

// preconditions returns false when the request's If-Match doesn't have the
// content's etag, by strong comparison, and isn't *, or without If-Match,
// the content has been modified since If-Unmodified-Since, or for a request
// other than GET or HEAD, when its If-None-Match has the content's etag
func (b Binary) preconditions(r *http.Request) bool {
	if match := r.Header.Get("If-Match"); match != "" {
		if !b.matches(match, true) {
			return false
		}
	} else if since, err := http.ParseTime(r.Header.Get("If-Unmodified-Since")); err == nil && !b.ModTime.IsZero() {
		if b.ModTime.Truncate(time.Second).After(since) {
			return false
		}
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		if match := r.Header.Get("If-None-Match"); match != "" && b.matches(match, false) {
			return false
		}
	}
	return true
}

// matches returns true when the list of entity tags in a header, such as
// If-None-Match, is *, or has the content's etag. With strong comparison, a
// weak tag doesn't match. A list which can't be parsed doesn't match.
func (b Binary) matches(value string, strong bool) bool {
	tags, all, ok := parseETags(value)
	if !ok {
		return false
	} else if all {
		return true
	}
	for _, tag := range tags {
		if b.ETag != "" && tag.opaque == b.ETag && !(strong && tag.weak) {
			return true
		}
	}
	return false
}

// entityTag is an entity tag of a request's header, without its quotes,
// which is weak when it starts with W/
type entityTag struct {
	opaque string
	weak   bool
}

// parseETags returns the entity tags in a list of them, such as
// "a", W/"b", whose quoted tags can have commas, or all when the list is *,
// which matches every tag. Returns false when the list can't be parsed.
func parseETags(value string) (tags []entityTag, all bool, ok bool) {
	value = strings.TrimSpace(value)
	if value == "*" {
		return nil, true, true
	}
	for {
		// Skip the separators, which can be empty elements
		value = strings.TrimLeft(value, " \t,")
		if value == "" {
			return tags, false, len(tags) > 0
		}

		// A tag is weak with W/, and is quoted, without quotes in it
		var tag entityTag
		if rest, found := strings.CutPrefix(value, "W/"); found {
			tag.weak, value = true, rest
		}
		if !strings.HasPrefix(value, `"`) {
			return nil, false, false
		}
		end := strings.IndexByte(value[1:], '"')
		if end < 0 {
			return nil, false, false
		}
		tag.opaque, value = value[1:1+end], value[2+end:]
		tags = append(tags, tag)

		// A tag is followed by a separator, or the end of the list
		value = strings.TrimLeft(value, " \t")
		if value != "" && value[0] != ',' {
			return nil, false, false
		}
	}
}

// headerJSON returns the value as JSON for a header, with the characters
// which aren't ASCII escaped as \uXXXX, so it's ASCII, as browsers decode
// headers as Latin-1, which would garble UTF-8. The escaped JSON decodes to
// the same value.
func headerJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// Characters which aren't ASCII are only in strings, where they can be
	// escaped, and those beyond the Basic Multilingual Plane are escaped as
	// UTF-16 surrogate pairs, as JSON requires
	var result strings.Builder
	result.Grow(len(data))
	for _, r := range string(data) {
		switch {
		case r < 0x80:
			result.WriteRune(r)
		case r > 0xFFFF:
			r1, r2 := utf16.EncodeRune(r)
			fmt.Fprintf(&result, `\u%04x\u%04x`, r1, r2)
		default:
			fmt.Fprintf(&result, `\u%04x`, r)
		}
	}
	return result.String(), nil
}

// scriptable returns true when content of the type can run scripts in a
// browser, which is HTML and XML, including SVG, or a type which can't be
// parsed, to be safe
func scriptable(contentType string) bool {
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return true
	}
	switch {
	case mediatype == "text/html", mediatype == "application/xhtml+xml":
		return true
	case strings.HasSuffix(mediatype, "/xml"), strings.HasSuffix(mediatype, "+xml"):
		return true
	default:
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////
// RECORDERS

// readSeekRecorder records the first error reading or seeking, other than
// the end of the content
type readSeekRecorder struct {
	io.ReadSeeker
	err error
}

func (r *readSeekRecorder) Read(p []byte) (int, error) {
	n, err := r.ReadSeeker.Read(p)
	if err != nil && !errors.Is(err, io.EOF) && r.err == nil {
		r.err = err
	}
	return n, err
}

func (r *readSeekRecorder) Seek(offset int64, whence int) (int64, error) {
	n, err := r.ReadSeeker.Seek(offset, whence)
	if err != nil && r.err == nil {
		r.err = err
	}
	return n, err
}

// writeRecorder records the first error writing a response, and unwraps to
// it, for http.ResponseController
type writeRecorder struct {
	http.ResponseWriter
	err error
}

func (w *writeRecorder) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if err != nil && w.err == nil {
		w.err = err
	}
	return n, err
}

func (w *writeRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
