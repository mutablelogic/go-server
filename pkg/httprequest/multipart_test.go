package httprequest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

// buildForm creates an in-memory multipart.Form with one file part so we have
// a real *multipart.FileHeader to pass to multipartCleanup.open.
func buildForm(t *testing.T, filename, content string) *multipart.Form {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(content))
	w.Close()
	r := multipart.NewReader(&buf, w.Boundary())
	form, err := r.ReadForm(32 << 20)
	if err != nil {
		t.Fatal(err)
	}
	return form
}

func Test_multipartCleanup_NoFiles(t *testing.T) {
	assert := assert.New(t)
	// With no files opened, done() should trigger RemoveAll immediately
	// without panicking or decrementing below zero.
	form := buildForm(t, "a.txt", "data")
	mc := &multipartCleanup{form: form}
	mc.done()
	assert.Equal(int32(0), mc.refs.Load())
}

func Test_multipartCleanup_SingleFile_CloseTriggersRemoveAll(t *testing.T) {
	assert := assert.New(t)
	form := buildForm(t, "a.txt", "hello")
	mc := &multipartCleanup{form: form}
	fh := form.File["file"][0]
	rc, err := mc.open(fh)
	assert.NoError(err)
	assert.Equal(int32(1), mc.refs.Load())
	// done() should be a no-op since refs > 0.
	mc.done()
	assert.Equal(int32(1), mc.refs.Load())
	// Read and close; refs should drop to 0 and RemoveAll fire.
	data, _ := io.ReadAll(rc)
	assert.Equal("hello", string(data))
	assert.NoError(rc.Close())
	assert.Equal(int32(0), mc.refs.Load())
}

func Test_multipartCleanup_MultipleFiles_RemoveAllOnLastClose(t *testing.T) {
	assert := assert.New(t)
	// Build a form with two file parts under the same field name.
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	p1, _ := w.CreateFormFile("file", "a.txt")
	_, _ = p1.Write([]byte("aaa"))
	p2, _ := w.CreateFormFile("file", "b.txt")
	_, _ = p2.Write([]byte("bbb"))
	w.Close()
	r := multipart.NewReader(&buf, w.Boundary())
	form, err := r.ReadForm(32 << 20)
	assert.NoError(err)
	mc := &multipartCleanup{form: form}
	fhs := form.File["file"]
	assert.Len(fhs, 2)
	rc1, err := mc.open(fhs[0])
	assert.NoError(err)
	rc2, err := mc.open(fhs[1])
	assert.NoError(err)
	assert.Equal(int32(2), mc.refs.Load())
	// Close first; refs should still be 1.
	assert.NoError(rc1.Close())
	assert.Equal(int32(1), mc.refs.Load())
	// Close second; refs should hit 0 and RemoveAll fire.
	assert.NoError(rc2.Close())
	assert.Equal(int32(0), mc.refs.Load())
}

func Test_refCountedBody_DoubleCloseIsIdempotent(t *testing.T) {
	assert := assert.New(t)
	form := buildForm(t, "x.txt", "x")
	mc := &multipartCleanup{form: form}
	fh := form.File["file"][0]
	rc, err := mc.open(fh)
	assert.NoError(err)
	assert.Equal(int32(1), mc.refs.Load())
	assert.NoError(rc.Close())
	assert.Equal(int32(0), mc.refs.Load())
	// Second close must not decrement below zero.
	assert.NoError(rc.Close())
	assert.Equal(int32(0), mc.refs.Load())
}

func Test_multipartCleanup_OpenError(t *testing.T) {
	assert := assert.New(t)
	// A zero-value FileHeader has no content; Open() will fail.
	form := buildForm(t, "x.txt", "x")
	mc := &multipartCleanup{form: form}
	_, err := mc.open(&multipart.FileHeader{Filename: "broken.txt"})
	assert.Error(err)
	// refs must not have been incremented on failure.
	assert.Equal(int32(0), mc.refs.Load())
}

// buildRequest creates a multipart/form-data POST request from a pre-built
// multipart.Writer for use in Read integration tests.
func buildRequest(buf *bytes.Buffer, w *multipart.Writer) *http.Request {
	r, _ := http.NewRequest(http.MethodPost, "/", buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}

func Test_Read_FormData_Files(t *testing.T) {
	assert := assert.New(t)

	t.Run("WithFileField", func(t *testing.T) {
		type payload struct {
			Name string     `json:"name"`
			File types.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("name", "doc")
		part, _ := w.CreateFormFile("file", "test.txt")
		_, _ = part.Write([]byte("file contents"))
		w.Close()
		var p payload
		err := Read(buildRequest(&buf, w), &p)
		assert.NoError(err)
		assert.Equal("doc", p.Name)
		assert.Equal("test.txt", p.File.Path)
		assert.NotNil(p.File.Body)
		assert.NotEmpty(p.File.ContentType)
		assert.NotNil(p.File.Header)
		assert.NotEmpty(p.File.Header.Get("Content-Disposition"))
		defer p.File.Body.Close()
		data, _ := io.ReadAll(p.File.Body)
		assert.Equal("file contents", string(data))
	})

	t.Run("UnsupportedFileFieldType", func(t *testing.T) {
		type payload struct {
			File string `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, _ := w.CreateFormFile("file", "test.txt")
		_, _ = part.Write([]byte("data"))
		w.Close()
		var p payload
		err := Read(buildRequest(&buf, w), &p)
		assert.Error(err)
	})

	t.Run("WithMultipleFilesSlice", func(t *testing.T) {
		type payload struct {
			Name  string       `json:"name"`
			Files []types.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("name", "batch")
		part1, _ := w.CreateFormFile("file", "a.txt")
		_, _ = part1.Write([]byte("contents of a"))
		part2, _ := w.CreateFormFile("file", "b.txt")
		_, _ = part2.Write([]byte("contents of b"))
		w.Close()
		var p payload
		err := Read(buildRequest(&buf, w), &p)
		assert.NoError(err)
		assert.Equal("batch", p.Name)
		assert.Len(p.Files, 2)
		assert.Equal("a.txt", p.Files[0].Path)
		assert.Equal("b.txt", p.Files[1].Path)
		assert.NotEmpty(p.Files[0].Header.Get("Content-Disposition"))
		assert.NotEmpty(p.Files[1].Header.Get("Content-Disposition"))
		assert.NotEmpty(p.Files[0].ContentType)
		assert.NotEmpty(p.Files[1].ContentType)
		defer p.Files[0].Body.Close()
		defer p.Files[1].Body.Close()
		data0, _ := io.ReadAll(p.Files[0].Body)
		data1, _ := io.ReadAll(p.Files[1].Body)
		assert.Equal("contents of a", string(data0))
		assert.Equal("contents of b", string(data1))
	})

	t.Run("WithSingleFileInSlice", func(t *testing.T) {
		type payload struct {
			Files []types.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, _ := w.CreateFormFile("file", "only.txt")
		_, _ = part.Write([]byte("solo"))
		w.Close()
		var p payload
		err := Read(buildRequest(&buf, w), &p)
		assert.NoError(err)
		assert.Len(p.Files, 1)
		assert.Equal("only.txt", p.Files[0].Path)
		defer p.Files[0].Body.Close()
		data, _ := io.ReadAll(p.Files[0].Body)
		assert.Equal("solo", string(data))
	})

	t.Run("XPathHeader_SingleFile", func(t *testing.T) {
		// Verify that when a part carries an X-Path header the resulting
		// File.Path equals the X-Path value rather than the basename.
		type payload struct {
			File types.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, "file", "z.txt"))
		h.Set("Content-Type", "text/plain")
		h.Set(types.ContentPathHeader, "x/y/z.txt")
		part, err := w.CreatePart(h)
		assert.NoError(err)
		_, _ = part.Write([]byte("deep content"))
		w.Close()
		var p payload
		err = Read(buildRequest(&buf, w), &p)
		assert.NoError(err)
		assert.Equal("x/y/z.txt", p.File.Path)
		defer p.File.Body.Close()
		data, _ := io.ReadAll(p.File.Body)
		assert.Equal("deep content", string(data))
	})

	t.Run("XPathHeader_FileSlice", func(t *testing.T) {
		// Verify X-Path is respected for every part when binding to []types.File.
		type payload struct {
			Files []types.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		for _, tc := range []struct{ xpath, content string }{
			{"a/b/one.txt", "content one"},
			{"c/d/two.txt", "content two"},
		} {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, "file", tc.xpath[len(tc.xpath)-7:]))
			h.Set("Content-Type", "text/plain")
			h.Set(types.ContentPathHeader, tc.xpath)
			part, err := w.CreatePart(h)
			assert.NoError(err)
			_, _ = part.Write([]byte(tc.content))
		}
		w.Close()
		var p payload
		err := Read(buildRequest(&buf, w), &p)
		assert.NoError(err)
		assert.Len(p.Files, 2)
		assert.Equal("a/b/one.txt", p.Files[0].Path)
		assert.Equal("c/d/two.txt", p.Files[1].Path)
		defer p.Files[0].Body.Close()
		defer p.Files[1].Body.Close()
		data0, _ := io.ReadAll(p.Files[0].Body)
		data1, _ := io.ReadAll(p.Files[1].Body)
		assert.Equal("content one", string(data0))
		assert.Equal("content two", string(data1))
	})

	t.Run("FileSliceOpenError", func(t *testing.T) {
		// Verify that when a file in the slice cannot be opened, an error is
		// returned and previously-opened bodies are closed (no resource leak).
		type payload struct {
			Files []types.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, _ := w.CreateFormFile("file", "good.txt")
		_, _ = part.Write([]byte("data"))
		w.Close()
		r := buildRequest(&buf, w)
		// Pre-parse so we can tamper with the file list.
		_ = r.ParseMultipartForm(32 << 20)
		// Append a zero-value FileHeader: Open() will fail.
		r.MultipartForm.File["file"] = append(
			r.MultipartForm.File["file"],
			&multipart.FileHeader{Filename: "broken.txt"},
		)
		var p payload
		err := Read(r, &p)
		assert.Error(err)
	})
}
