package types

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
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
	ContentMetaHeader        = "X-Meta"
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

// AcceptsContentType returns true when the request accepts the content type,
// such as text/event-stream, by its Accept headers, which it does when it
// has none. Each of the media ranges is matched, including those with a
// wildcard, such as text/* or */*, and the most specific which matches
// decides, by its quality, q, which doesn't accept it when it's 0. Returns
// an error when a media range, or its quality, can't be parsed.
func AcceptsContentType(r *http.Request, contentType string) (bool, error) {
	typ, subtype, _ := strings.Cut(strings.ToLower(contentType), "/")
	matched, quality := 0, 0.0
	for _, header := range r.Header.Values(ContentAcceptHeader) {
		for _, value := range strings.Split(header, ",") {
			if value = strings.TrimSpace(value); value == "" {
				continue
			}
			mediaRange, params, err := mime.ParseMediaType(value)
			if err != nil {
				return false, fmt.Errorf("accept %q: %w", value, err)
			} else if !strings.Contains(mediaRange, "/") {
				return false, fmt.Errorf("accept %q: not a media range", value)
			}
			q := 1.0
			if value, exists := params["q"]; exists {
				if q, err = strconv.ParseFloat(value, 64); err != nil || q < 0 || q > 1 {
					return false, fmt.Errorf("accept %q: invalid quality %q", mediaRange, value)
				}
			}

			// The most specific media range which matches decides
			rangeType, rangeSubtype, _ := strings.Cut(mediaRange, "/")
			var specificity int
			switch {
			case rangeType == typ && rangeSubtype == subtype:
				specificity = 3
			case rangeType == typ && rangeSubtype == "*":
				specificity = 2
			case rangeType == "*" && rangeSubtype == "*":
				specificity = 1
			default:
				continue
			}
			if specificity > matched {
				matched, quality = specificity, q
			}
		}
	}

	// Accept the content type when there are no Accept headers, or the most
	// specific media range which matches has a quality
	if len(r.Header.Values(ContentAcceptHeader)) == 0 {
		return true, nil
	}
	return matched > 0 && quality > 0, nil
}
