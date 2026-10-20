package httprequest_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// formPart is a part of a form, which is a file when it has a file name
type formPart struct {
	field, filename, value string
	header                 textproto.MIMEHeader
}

// formRequest returns a request with a multipart form of the parts
func formRequest(t *testing.T, parts ...formPart) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		header := textproto.MIMEHeader{}
		for key, values := range part.header {
			header[key] = values
		}
		disposition := `form-data; name="` + part.field + `"`
		if part.filename != "" {
			disposition += `; filename="` + part.filename + `"`
		}
		header.Set("Content-Disposition", disposition)
		w, err := writer.CreatePart(header)
		require.NoError(t, err)
		_, err = w.Write([]byte(part.value))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// upload is a file which was read, with the values before it
type upload struct {
	field   string
	values  url.Values
	path    string
	content string
	ctype   string
	modTime time.Time
}

// readFiles reads the files of the request, and returns them
func readFiles(t *testing.T, req *http.Request) ([]upload, url.Values, error) {
	t.Helper()
	var uploads []upload
	values, err := httprequest.Binary(req, func(field string, values url.Values, file types.File) error {
		data, err := io.ReadAll(file.Body)
		if err != nil {
			return err
		}
		uploads = append(uploads, upload{field, cloneValues(values), file.Path, string(data), file.ContentType, file.ModTime})
		return nil
	})
	return uploads, values, err
}

func cloneValues(values url.Values) url.Values {
	result := url.Values{}
	for key, vs := range values {
		result[key] = append([]string(nil), vs...)
	}
	return result
}

func Test_Binary(t *testing.T) {
	mtime := time.Date(2014, 3, 4, 5, 42, 8, 0, time.UTC)

	t.Run("Body", func(t *testing.T) {
		// A request which isn't a form is one file, with its type and
		// modified time
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("hello, world"))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Last-Modified", mtime.Format(http.TimeFormat))
		uploads, values, err := readFiles(t, req)
		require.NoError(t, err)
		require.Len(t, uploads, 1)
		assert.Equal(t, "", uploads[0].field)
		assert.Equal(t, "", uploads[0].path)
		assert.Equal(t, "hello, world", uploads[0].content)
		assert.Equal(t, "text/plain", uploads[0].ctype)
		assert.True(t, uploads[0].modTime.Equal(mtime))
		assert.Empty(t, values)
	})

	t.Run("Form", func(t *testing.T) {
		// The values before a file are passed with it, and all of them are
		// returned
		req := formRequest(t,
			formPart{field: "session", value: "abc"},
			formPart{field: "file", filename: "a.txt", value: "first", header: textproto.MIMEHeader{"Content-Type": {"text/plain"}}},
			formPart{field: "type", value: "text/markdown"},
			formPart{field: "file", filename: "b.md", value: "second"},
			formPart{field: "after", value: "done"},
		)
		uploads, values, err := readFiles(t, req)
		require.NoError(t, err)
		require.Len(t, uploads, 2)
		assert.Equal(t, upload{field: "file", values: url.Values{"session": {"abc"}}, path: "a.txt", content: "first", ctype: "text/plain"}, uploads[0])
		assert.Equal(t, url.Values{"session": {"abc"}, "type": {"text/markdown"}}, uploads[1].values)
		assert.Equal(t, "b.md", uploads[1].path)
		assert.Equal(t, url.Values{"session": {"abc"}, "type": {"text/markdown"}, "after": {"done"}}, values)
	})

	t.Run("Paths", func(t *testing.T) {
		// A file's path keeps its directories, from its X-Path, or its file
		// name, as a browser sends a folder's files, and can't be above its
		// root
		for want, part := range map[string]formPart{
			"Photos/2024/a.jpg": {field: "file", filename: "Photos/2024/a.jpg"},
			"Photos/b.jpg":      {field: "file", filename: `Photos\b.jpg`},
			"x/y/z.txt":         {field: "file", filename: "z.txt", header: textproto.MIMEHeader{types.ContentPathHeader: {"/x/y/z.txt"}}},
			"passwd":            {field: "file", filename: "../../../etc/passwd"},
			"c.txt":             {field: "file", filename: "c.txt", header: textproto.MIMEHeader{types.ContentPathHeader: {"../../c.txt"}}},
			"d.txt":             {field: "file", filename: "/d.txt"},
		} {
			uploads, _, err := readFiles(t, formRequest(t, part))
			require.NoError(t, err, want)
			require.Len(t, uploads, 1, want)
			assert.Equal(t, want, uploads[0].path, "%+v", part)
		}
	})

	t.Run("ModTime", func(t *testing.T) {
		req := formRequest(t, formPart{field: "file", filename: "a.txt", value: "x", header: textproto.MIMEHeader{"Last-Modified": {mtime.Format(http.TimeFormat)}}})
		uploads, _, err := readFiles(t, req)
		require.NoError(t, err)
		require.Len(t, uploads, 1)
		assert.True(t, uploads[0].modTime.Equal(mtime))
	})

	t.Run("Unread", func(t *testing.T) {
		// A file which fn doesn't read is skipped
		req := formRequest(t,
			formPart{field: "file", filename: "a.txt", value: strings.Repeat("a", 100000)},
			formPart{field: "file", filename: "b.txt", value: "second"},
		)
		var paths []string
		_, err := httprequest.Binary(req, func(_ string, _ url.Values, file types.File) error {
			paths = append(paths, file.Path)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"a.txt", "b.txt"}, paths)
	})

	t.Run("Errors", func(t *testing.T) {
		// An error from fn stops reading, and is returned
		failed := errors.New("failed")
		req := formRequest(t, formPart{field: "file", filename: "a.txt"}, formPart{field: "file", filename: "b.txt"})
		calls := 0
		_, err := httprequest.Binary(req, func(string, url.Values, types.File) error {
			calls++
			return failed
		})
		assert.ErrorIs(t, err, failed)
		assert.Equal(t, 1, calls)

		// A value which is too large is a bad request
		_, _, err = readFiles(t, formRequest(t, formPart{field: "big", value: strings.Repeat("x", httprequest.FormValueMaxSize+1)}))
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)

		// A file which is cut off is an error reading it, in fn
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("--xyz\r\nContent-Disposition: form-data; name=\"file\"; filename=\"a.txt\"\r\n\r\npartial"))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=xyz")
		_, _, err = readFiles(t, req)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)

		// A form whose parts can't be read is a bad request
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("--xyz\r\nnot a header\r\n\r\nvalue\r\n--xyz--\r\n"))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=xyz")
		_, _, err = readFiles(t, req)
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)

		// A form without a boundary is a bad request
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("value"))
		req.Header.Set("Content-Type", "multipart/form-data")
		_, _, err = readFiles(t, req)
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)
	})
}

func Test_Read_FilePath(t *testing.T) {
	// Read keeps a browser's folder path, and the modified time, as
	// Binary does
	mtime := time.Date(2014, 3, 4, 5, 42, 8, 0, time.UTC)
	var form struct {
		File types.File `json:"file"`
	}
	req := formRequest(t, formPart{field: "file", filename: "Photos/2024/a.jpg", value: "x", header: textproto.MIMEHeader{"Last-Modified": {mtime.Format(http.TimeFormat)}}})
	require.NoError(t, httprequest.Read(req, &form))
	defer form.File.Body.Close()
	assert.Equal(t, "Photos/2024/a.jpg", form.File.Path)
	assert.True(t, form.File.ModTime.Equal(mtime))
}
