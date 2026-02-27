package httprequest_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	gomultipart "github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/stretchr/testify/assert"
)

func Test_Read_JSON(t *testing.T) {
	assert := assert.New(t)

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("Valid", func(t *testing.T) {
		body := strings.NewReader(`{"name":"alice","age":30}`)
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json")
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
		assert.Equal(30, p.Age)
	})

	t.Run("EmptyBody", func(t *testing.T) {
		body := strings.NewReader("")
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json")
		var p payload
		err := httprequest.Read(r, &p)
		assert.Error(err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := strings.NewReader("{invalid")
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json")
		var p payload
		err := httprequest.Read(r, &p)
		assert.Error(err)
	})
}

func Test_Read_TextPlain_String(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("hello world")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var s string
	err := httprequest.Read(r, &s)
	assert.NoError(err)
	assert.Equal("hello world", s)
}

func Test_Read_TextPlain_Bytes(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("binary data")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var b []byte
	err := httprequest.Read(r, &b)
	assert.NoError(err)
	assert.Equal([]byte("binary data"), b)
}

func Test_Read_TextPlain_UnsupportedType(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("hello")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var i int
	err := httprequest.Read(r, &i)
	assert.Error(err)
}

func Test_Read_NoContentType(t *testing.T) {
	assert := assert.New(t)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	var s string
	err := httprequest.Read(r, &s)
	assert.Error(err)
}

func Test_Read_UnsupportedContentType(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("<xml/>")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "application/xml")
	var s string
	err := httprequest.Read(r, &s)
	assert.Error(err)
}

func Test_Indent(t *testing.T) {
	assert := assert.New(t)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(2, httprequest.Indent(r))
}

func Test_Read_FormData(t *testing.T) {
	assert := assert.New(t)

	t.Run("WithFields", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
			Age  string `json:"age"`
		}
		// Build multipart body
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("name", "alice")
		_ = w.WriteField("age", "30")
		w.Close()

		r, _ := http.NewRequest(http.MethodPost, "/", &buf)
		r.Header.Set("Content-Type", w.FormDataContentType())
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
		assert.Equal("30", p.Age)
	})

	t.Run("WithFileField", func(t *testing.T) {
		type payload struct {
			Name string           `json:"name"`
			File gomultipart.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("name", "doc")
		part, _ := w.CreateFormFile("file", "test.txt")
		_, _ = part.Write([]byte("file contents"))
		w.Close()

		r, _ := http.NewRequest(http.MethodPost, "/", &buf)
		r.Header.Set("Content-Type", w.FormDataContentType())
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("doc", p.Name)
		assert.Equal("test.txt", p.File.Path)
		assert.NotNil(p.File.Body)
		// Read body
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

		r, _ := http.NewRequest(http.MethodPost, "/", &buf)
		r.Header.Set("Content-Type", w.FormDataContentType())
		var p payload
		err := httprequest.Read(r, &p)
		assert.Error(err)
	})

	t.Run("WithMultipleFilesSlice", func(t *testing.T) {
		type payload struct {
			Name  string             `json:"name"`
			Files []gomultipart.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("name", "batch")
		part1, _ := w.CreateFormFile("file", "a.txt")
		_, _ = part1.Write([]byte("contents of a"))
		part2, _ := w.CreateFormFile("file", "b.txt")
		_, _ = part2.Write([]byte("contents of b"))
		w.Close()

		r, _ := http.NewRequest(http.MethodPost, "/", &buf)
		r.Header.Set("Content-Type", w.FormDataContentType())
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("batch", p.Name)
		assert.Len(p.Files, 2)
		assert.Equal("a.txt", p.Files[0].Path)
		assert.Equal("b.txt", p.Files[1].Path)
		data0, _ := io.ReadAll(p.Files[0].Body)
		data1, _ := io.ReadAll(p.Files[1].Body)
		assert.Equal("contents of a", string(data0))
		assert.Equal("contents of b", string(data1))
	})

	t.Run("WithSingleFileInSlice", func(t *testing.T) {
		type payload struct {
			Files []gomultipart.File `json:"file"`
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, _ := w.CreateFormFile("file", "only.txt")
		_, _ = part.Write([]byte("solo"))
		w.Close()

		r, _ := http.NewRequest(http.MethodPost, "/", &buf)
		r.Header.Set("Content-Type", w.FormDataContentType())
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Len(p.Files, 1)
		assert.Equal("only.txt", p.Files[0].Path)
		data, _ := io.ReadAll(p.Files[0].Body)
		assert.Equal("solo", string(data))
	})
}

func Test_Read_FormData_JSON(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSONContentTypeWithCharset", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
		}
		body := strings.NewReader(`{"name":"bob"}`)
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("bob", p.Name)
	})
}
