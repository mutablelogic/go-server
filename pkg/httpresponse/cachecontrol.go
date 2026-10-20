package httpresponse

import (
	"strconv"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// CacheControl is how a response is cached, in its Cache-Control header.
// The zero value is private, no-cache: a browser can keep the response, but
// must check it hasn't changed before using it, and a shared cache, such as
// a proxy, doesn't keep it, which is safe for content which can change, or
// needs authentication.
type CacheControl struct {
	// Public is true when a shared cache, such as a proxy or a CDN, can keep
	// the response, rather than only the browser
	Public bool

	// MaxAge is how long the response is used without checking it hasn't
	// changed, to the second, or when it's zero, it's always checked
	MaxAge time.Duration

	// Immutable is true when the response never changes, so it isn't checked
	// when the page is reloaded, while it's within MaxAge, such as content
	// whose etag is its hash
	Immutable bool

	// NoStore is true when the response isn't kept at all, such as sensitive
	// data, which overrides the others
	NoStore bool
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// String returns the value of the Cache-Control header
func (c CacheControl) String() string {
	if c.NoStore {
		return "no-store"
	}
	result := "private"
	if c.Public {
		result = "public"
	}
	if seconds := int64(c.MaxAge / time.Second); seconds > 0 {
		result += ", max-age=" + strconv.FormatInt(seconds, 10)
		if c.Immutable {
			result += ", immutable"
		}
	} else {
		result += ", no-cache"
	}
	return result
}
