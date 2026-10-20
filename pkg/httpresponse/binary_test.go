package httpresponse_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// onlyReader hides the other methods of a reader, such as Seek
type onlyReader struct{ io.Reader }

// slowReader returns its data a byte at a time, after a delay
type slowReader struct {
	data  string
	delay time.Duration
}

func (r *slowReader) Read(p []byte) (int, error) {
	if r.data == "" {
		return 0, io.EOF
	}
	time.Sleep(r.delay)
	n := copy(p[:1], r.data)
	r.data = r.data[n:]
	return n, nil
}

func Test_Binary(t *testing.T) {
	mtime := time.Date(2014, 3, 4, 5, 42, 8, 500_000_000, time.UTC)
	content := "hello, world"
	binary := httpresponse.Binary{
		ContentType:  "text/plain; charset=utf-8",
		Size:         types.Ptr(int64(len(content))),
		ModTime:      mtime,
		ETag:         "abc",
		Path:         "/docs/hello.txt",
		Disposition:  httpresponse.DispositionInline,
		CacheControl: httpresponse.CacheControl{MaxAge: time.Hour},
		Meta:         map[string]string{"title": "Café"},
	}

	// write responds to the request, with the headers, with the content
	write := func(t *testing.T, binary httpresponse.Binary, method string, header http.Header, body io.Reader) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "/hello.txt", nil)
		for key, values := range header {
			req.Header[key] = values
		}
		rec := httptest.NewRecorder()
		require.NoError(t, binary.Write(rec, req, body))
		return rec
	}

	t.Run("Get", func(t *testing.T) {
		rec := write(t, binary, http.MethodGet, nil, strings.NewReader(content))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, content, rec.Body.String())
		assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "12", rec.Header().Get("Content-Length"))
		assert.Equal(t, "Tue, 04 Mar 2014 05:42:08 GMT", rec.Header().Get("Last-Modified"))
		assert.Equal(t, `"abc"`, rec.Header().Get("ETag"))
		assert.Equal(t, "private, max-age=3600", rec.Header().Get("Cache-Control"))

		// Without a cache control, it's always checked
		other := binary
		other.CacheControl = httpresponse.CacheControl{}
		assert.Equal(t, "private, no-cache", write(t, other, http.MethodHead, nil, nil).Header().Get("Cache-Control"))
		assert.Equal(t, "inline; filename=hello.txt", rec.Header().Get("Content-Disposition"))
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
	})

	t.Run("Meta", func(t *testing.T) {
		// The metadata is ASCII, and decodes to the same value
		rec := write(t, binary, http.MethodHead, nil, nil)
		meta := rec.Header().Get(types.ContentMetaHeader)
		assert.Contains(t, meta, `Caf\u00e9`)
		for _, b := range []byte(meta) {
			assert.Less(t, b, byte(0x80))
		}
		var decoded map[string]string
		require.NoError(t, json.Unmarshal([]byte(meta), &decoded))
		assert.Equal(t, "Café", decoded["title"])

		// Without metadata, there's no header
		other := binary
		other.Meta = nil
		assert.Empty(t, write(t, other, http.MethodHead, nil, nil).Header().Get(types.ContentMetaHeader))

		// Metadata which can't be marshalled is an error, and nothing is
		// written
		other.Meta = func() {}
		rec = httptest.NewRecorder()
		assert.Error(t, other.Write(rec, httptest.NewRequest(http.MethodGet, "/", nil), strings.NewReader(content)))
		assert.Empty(t, rec.Body.String())
	})

	t.Run("Disposition", func(t *testing.T) {
		// An attachment, with a name which isn't ASCII
		other := binary
		other.Disposition, other.Path = httpresponse.DispositionAttachment, "/docs/café.txt"
		assert.Equal(t, "attachment; filename*=utf-8''caf%C3%A9.txt", write(t, other, http.MethodHead, nil, nil).Header().Get("Content-Disposition"))

		// Without a path, there's no file name, and without a disposition,
		// there's no header
		other.Path = ""
		assert.Equal(t, "attachment", write(t, other, http.MethodHead, nil, nil).Header().Get("Content-Disposition"))
		other.Disposition, other.Path = "", "/docs/hello.txt"
		assert.Empty(t, write(t, other, http.MethodHead, nil, nil).Header().Get("Content-Disposition"))
	})

	t.Run("Head", func(t *testing.T) {
		// The headers alone
		rec := write(t, binary, http.MethodHead, nil, nil)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())
		assert.Equal(t, "12", rec.Header().Get("Content-Length"))
	})

	t.Run("NoBody", func(t *testing.T) {
		// A GET request without a body has no content
		rec := write(t, binary, http.MethodGet, nil, nil)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Length"))
	})

	t.Run("NotModified", func(t *testing.T) {
		for name, tt := range map[string]struct {
			header http.Header
			want   bool
		}{
			"ETag":         {http.Header{"If-None-Match": {`"abc"`}}, true},
			"WeakETag":     {http.Header{"If-None-Match": {`W/"abc"`}}, true},
			"ETags":        {http.Header{"If-None-Match": {`"def", "abc"`}}, true},
			"Any":          {http.Header{"If-None-Match": {"*"}}, true},
			"OtherETag":    {http.Header{"If-None-Match": {`"def"`}}, false},
			"Since":        {http.Header{"If-Modified-Since": {"Tue, 04 Mar 2014 05:42:08 GMT"}}, true},
			"SinceLater":   {http.Header{"If-Modified-Since": {"Wed, 05 Mar 2014 05:42:08 GMT"}}, true},
			"SinceEarlier": {http.Header{"If-Modified-Since": {"Tue, 04 Mar 2014 05:42:07 GMT"}}, false},
			"ETagPrecedes": {http.Header{"If-None-Match": {`"def"`}, "If-Modified-Since": {"Wed, 05 Mar 2014 05:42:08 GMT"}}, false},
			"NoConditions": {nil, false},
			"InvalidSince": {http.Header{"If-Modified-Since": {"yesterday"}}, false},
		} {
			t.Run(name, func(t *testing.T) {
				rec := write(t, binary, http.MethodGet, tt.header, strings.NewReader(content))
				if tt.want {
					assert.Equal(t, http.StatusNotModified, rec.Code)
					assert.Empty(t, rec.Body.String())
					assert.Equal(t, `"abc"`, rec.Header().Get("ETag"))
				} else {
					assert.Equal(t, http.StatusOK, rec.Code)
					assert.Equal(t, content, rec.Body.String())
				}
			})
		}

		// Only GET and HEAD requests are not modified
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		req.Header.Set("If-None-Match", "*")
		assert.False(t, binary.NotModified(req))

		// Content without a modified time is modified since any time
		other := binary
		other.ModTime = time.Time{}
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("If-Modified-Since", "Wed, 05 Mar 2014 05:42:08 GMT")
		assert.False(t, other.NotModified(req))
	})

	t.Run("Range", func(t *testing.T) {
		// Content which can seek can be requested in ranges
		rec := write(t, binary, http.MethodGet, http.Header{"Range": {"bytes=7-"}}, strings.NewReader(content))
		assert.Equal(t, http.StatusPartialContent, rec.Code)
		assert.Equal(t, "world", rec.Body.String())
		assert.Equal(t, "5", rec.Header().Get("Content-Length"))

		// Content which can't is written whole
		rec = write(t, binary, http.MethodGet, http.Header{"Range": {"bytes=7-"}}, onlyReader{strings.NewReader(content)})
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, content, rec.Body.String())
	})

	t.Run("Sandboxed", func(t *testing.T) {
		for contentType, sandboxed := range map[string]bool{
			"text/html; charset=utf-8": true,
			"application/xhtml+xml":    true,
			"image/svg+xml":            true,
			"application/xml":          true,
			"text/xml":                 true,
			"not a type":               true,
			"video/mp4":                false,
			"application/pdf":          false,
			"image/jpeg":               false,
			"text/plain":               false,
			"":                         false,
		} {
			other := binary
			other.ContentType = contentType
			rec := write(t, other, http.MethodHead, nil, nil)
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"), contentType)
			if sandboxed {
				assert.Equal(t, "sandbox", rec.Header().Get("Content-Security-Policy"), contentType)
			} else {
				assert.Empty(t, rec.Header().Get("Content-Security-Policy"), contentType)
			}
		}

		// Without a type, it's binary
		other := binary
		other.ContentType = ""
		assert.Equal(t, types.ContentTypeBinary, write(t, other, http.MethodHead, nil, nil).Header().Get("Content-Type"))
	})

	t.Run("WriteTimeout", func(t *testing.T) {
		// Content which takes longer to write than the server's write
		// timeout is written whole
		slow := "slow content"
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			other := binary
			other.Size = types.Ptr(int64(len(slow)))
			other.Write(w, r, onlyReader{&slowReader{data: slow, delay: 20 * time.Millisecond}})
		}))
		server.Config.WriteTimeout = 100 * time.Millisecond
		server.Start()
		defer server.Close()

		resp, err := server.Client().Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, slow, string(data))
	})
}
