package types_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_RequestContentType(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSON content type", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "application/json")
		ct, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("application/json", ct)
	})

	t.Run("Content type with charset", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "text/html; charset=utf-8")
		ct, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("text/html", ct)
	})

	t.Run("Empty content type", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		mimetype, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("application/octet-stream", mimetype)
	})

	t.Run("Multipart form data", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "multipart/form-data; boundary=abc123")
		ct, err := types.RequestContentType(r)
		assert.NoError(err)
		assert.Equal("multipart/form-data", ct)
	})
}

func Test_AcceptContentType(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSON accept", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept", "application/json")
		ct, err := types.AcceptContentType(r)
		assert.NoError(err)
		assert.Equal("application/json", ct)
	})

	t.Run("Wildcard accept", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept", "*/*")
		ct, err := types.AcceptContentType(r)
		assert.NoError(err)
		assert.Equal("*/*", ct)
	})

	t.Run("Empty accept", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		mimetype, err := types.AcceptContentType(r)
		assert.NoError(err)
		assert.Equal("", mimetype)
	})
}

func Test_AcceptsContentType(t *testing.T) {
	accepts := func(t *testing.T, contentType string, accept ...string) (bool, error) {
		t.Helper()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		for _, value := range accept {
			r.Header.Add(types.ContentAcceptHeader, value)
		}
		return types.AcceptsContentType(r, contentType)
	}
	for _, test := range []struct {
		accept []string
		want   bool
	}{
		{nil, true},
		{[]string{"text/event-stream"}, true},
		{[]string{"application/json, text/event-stream"}, true},
		{[]string{"application/json", "text/event-stream"}, true},
		{[]string{"text/*"}, true},
		{[]string{"*/*"}, true},
		{[]string{"application/json"}, false},
		{[]string{"text/event-stream;q=0"}, false},
		{[]string{"*/*, text/event-stream;q=0"}, false},
		{[]string{"text/*;q=0, text/event-stream;q=0.5"}, true},
		{[]string{"application/json, */*;q=0.1"}, true},
		{[]string{"TEXT/Event-Stream"}, true},
	} {
		got, err := accepts(t, types.ContentTypeTextStream, test.accept...)
		assert.NoError(t, err, test.accept)
		assert.Equal(t, test.want, got, test.accept)
	}

	// A media range, or quality, which can't be parsed is an error
	for _, accept := range []string{"text", "text/event-stream;q=x", "text/event-stream;q=2"} {
		_, err := accepts(t, types.ContentTypeTextStream, accept)
		assert.Error(t, err, accept)
	}
}
