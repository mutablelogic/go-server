package httpresponse_test

import (
	"testing"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	assert "github.com/stretchr/testify/assert"
)

func Test_CacheControl(t *testing.T) {
	year := 365 * 24 * time.Hour
	for _, tt := range []struct {
		cache httpresponse.CacheControl
		want  string
	}{
		// The zero value is always checked, and only kept by the browser
		{httpresponse.CacheControl{}, "private, no-cache"},
		{httpresponse.CacheControl{Public: true}, "public, no-cache"},

		// Used without checking for a time, to the second
		{httpresponse.CacheControl{MaxAge: time.Minute + 500*time.Millisecond}, "private, max-age=60"},
		{httpresponse.CacheControl{MaxAge: year, Immutable: true}, "private, max-age=31536000, immutable"},
		{httpresponse.CacheControl{Public: true, MaxAge: year, Immutable: true}, "public, max-age=31536000, immutable"},

		// Immutable needs a time, and less than a second is always checked
		{httpresponse.CacheControl{Immutable: true}, "private, no-cache"},
		{httpresponse.CacheControl{MaxAge: 500 * time.Millisecond}, "private, no-cache"},

		// Not kept, whatever else is set
		{httpresponse.CacheControl{NoStore: true, Public: true, MaxAge: year, Immutable: true}, "no-store"},
	} {
		assert.Equal(t, tt.want, tt.cache.String(), "%+v", tt.cache)
	}
}
