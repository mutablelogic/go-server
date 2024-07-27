package httprequest_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

func Test_body_00(t *testing.T) {
	assert := assert.New(t)

	t.Run("EOF", func(t *testing.T) {
		// We should receive an EOF error here - no body
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Content-Type", "application/json")
		err := httprequest.Body(nil, req)
		assert.ErrorIs(err, io.EOF)
	})

	t.Run("ContentType1", func(t *testing.T) {
		// We should receive a badparameter error - invalid content type
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Content-Type", "application/test")
		err := httprequest.Body(nil, req)
		assert.ErrorIs(err, ErrBadParameter)
	})

	t.Run("ContentType2", func(t *testing.T) {
		// We should receive a badparameter error - invalid content type
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Content-Type", "text/plain")
		err := httprequest.Body(nil, req, httpresponse.ContentTypeJSON)
		assert.ErrorIs(err, ErrBadParameter)
	})

	t.Run("Writer", func(t *testing.T) {
		var in, out bytes.Buffer
		in.WriteString("Hello, World!")

		// We should receive the text
		req := httptest.NewRequest("GET", "/", &in)
		req.Header.Set("Content-Type", "text/plain")
		err := httprequest.Body(&out, req)
		assert.NoError(err)
		assert.Equal("Hello, World!", out.String())
	})

	t.Run("Text", func(t *testing.T) {
		var in bytes.Buffer
		var out string
		in.WriteString("Hello, World!")

		// We should receive the text
		req := httptest.NewRequest("GET", "/", &in)
		req.Header.Set("Content-Type", "text/plain")
		err := httprequest.Body(&out, req)
		assert.NoError(err)
		assert.Equal("Hello, World!", out)
	})

	t.Run("JSON", func(t *testing.T) {
		var in bytes.Buffer
		var out string
		in.WriteString(strconv.Quote("Hello, World!"))

		// We should receive the JSON and decode it
		req := httptest.NewRequest("GET", "/", &in)
		req.Header.Set("Content-Type", "application/json")
		err := httprequest.Body(&out, req)
		assert.NoError(err)
		assert.Equal("Hello, World!", out)
	})

	t.Run("Form", func(t *testing.T) {
		values := url.Values{
			"key1": []string{"Hello, World!"},
		}
		var in bytes.Buffer
		var out struct {
			Key1 string `json:"key1"`
		}
		in.WriteString(values.Encode())

		// POST form data
		req := httptest.NewRequest("POST", "/", &in)
		req.Header.Set("Content-Type", httprequest.ContentTypeUrlEncoded)
		err := httprequest.Body(&out, req)
		assert.NoError(err)
		assert.Equal("Hello, World!", out.Key1)
	})

	t.Run("File", func(t *testing.T) {
		var in bytes.Buffer
		var out struct {
			File *multipart.FileHeader `json:"file"`
		}

		// instantiate multipart request
		multipartWriter := multipart.NewWriter(&in)

		// add form field
		filePart, _ := multipartWriter.CreateFormFile("file", "file.txt")
		filePart.Write([]byte("Hello, World!"))

		// Close the body - so it can be read
		multipartWriter.Close()

		// Create the request
		req := httptest.NewRequest(http.MethodPost, "/file", &in)
		req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

		// Parse the request body
		err := httprequest.Body(&out, req)
		assert.NoError(err)
		assert.NotNil(out.File)

		// Let's read the contents of the file
		file, err := out.File.Open()
		assert.NoError(err)
		defer file.Close()

		// Read the contents of the file
		data, err := io.ReadAll(file)
		assert.NoError(err)
		assert.Equal("Hello, World!", string(data))
	})
}
