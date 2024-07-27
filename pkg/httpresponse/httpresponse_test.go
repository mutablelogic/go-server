package httpresponse_test

import (
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

func Test_httpresponse_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("Empty_0", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Empty(resp, 200))
		assert.Equal(200, resp.Code)
		assert.Equal("", resp.Header().Get("Content-Type"))
		assert.Equal("", resp.Body.String())
	})

	t.Run("Empty_1", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Empty(resp, 200, "Foo", "Bar"))
		assert.Equal(200, resp.Code)
		assert.Equal("", resp.Header().Get("Content-Type"))
		assert.Equal("", resp.Body.String())
		assert.Equal("Bar", resp.Header().Get("Foo"))
	})
}

func Test_httpresponse_002(t *testing.T) {
	assert := assert.New(t)

	t.Run("Text_0", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Text(resp, "", 200))
		assert.Equal(200, resp.Code)
		assert.Equal("text/plain", resp.Header().Get("Content-Type"))
		assert.Equal("\n", resp.Body.String())
	})

	t.Run("Text_1", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Text(resp, "\n", 200))
		assert.Equal(200, resp.Code)
		assert.Equal("text/plain", resp.Header().Get("Content-Type"))
		assert.Equal("\n", resp.Body.String())
	})

	t.Run("Text_2", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.ErrorIs(httpresponse.Text(resp, "\n", 200, "Foo", "Bar", "Extra"), ErrBadParameter)
	})

	t.Run("Text_3", func(t *testing.T) {
		assert.ErrorIs(httpresponse.Text(nil, "\n", 200), ErrBadParameter)
	})

	t.Run("Text_4", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Text(resp, "hello world", 200, "Content-Type", "text/test"))
		assert.Equal(200, resp.Code)
		assert.Equal("text/test", resp.Header().Get("Content-Type"))
		assert.Equal("hello world\n", resp.Body.String())
	})
}

func Test_httpresponse_003(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSON_0", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.JSON(resp, "hello world", 200, 0))
		assert.Equal(200, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("\"hello world\"\n", resp.Body.String())
	})

	t.Run("JSON_1", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.JSON(resp, false, 200, 0, "Content-Type", "Bar"))
		assert.Equal(200, resp.Code)
		assert.Equal("Bar", resp.Header().Get("Content-Type"))
		assert.Equal("false\n", resp.Body.String())
	})

	t.Run("JSON_2", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.JSON(resp, []string{"a", "b"}, 200, 0))
		assert.Equal(200, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal(`["a","b"]`+"\n", resp.Body.String())
	})

	t.Run("JSON_3", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.JSON(resp, []string{"a", "b"}, 200, 1))
		assert.Equal(200, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("[\n \"a\",\n \"b\"\n]\n", resp.Body.String())
	})

	t.Run("JSON_4", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.JSON(resp, []string{"a", "b"}, 200, 2))
		assert.Equal(200, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("[\n  \"a\",\n  \"b\"\n]\n", resp.Body.String())
	})
}

func Test_httpresponse_004(t *testing.T) {
	assert := assert.New(t)

	t.Run("Error_0", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Error(resp, 404))
		assert.Equal(404, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("{\"code\":404,\"reason\":\"Not Found\"}\n", resp.Body.String())
	})

	t.Run("Error_1", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Error(resp, 404, "not found"))
		assert.Equal(404, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("{\"code\":404,\"reason\":\"not found\"}\n", resp.Body.String())
	})

	t.Run("Error_2", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Error(resp, 404, "not", "found"))
		assert.Equal(404, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("{\"code\":404,\"reason\":\"not found\"}\n", resp.Body.String())
	})

	t.Run("Error_3", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.ErrorWith(resp, 404, "this is the detail", "not", "found"))
		assert.Equal(404, resp.Code)
		assert.Equal("application/json", resp.Header().Get("Content-Type"))
		assert.Equal("{\n  \"code\": 404,\n  \"reason\": \"not found\",\n  \"detail\": \"this is the detail\"\n}\n", resp.Body.String())
	})
}

func Test_httpresponse_005(t *testing.T) {
	assert := assert.New(t)

	t.Run("Cors_0", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Cors(resp, "test"))
		assert.Equal(200, resp.Code)
		assert.Equal("test", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("Cors_1", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Cors(resp, ""))
		assert.Equal(200, resp.Code)
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("Cors_2", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Cors(resp, "", "get"))
		assert.Equal(200, resp.Code)
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal("GET", resp.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("Cors_3", func(t *testing.T) {
		resp := httptest.NewRecorder()
		assert.NoError(httpresponse.Cors(resp, "*", "get", "post"))
		assert.Equal(200, resp.Code)
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal("GET,POST", resp.Header().Get("Access-Control-Allow-Methods"))
	})
}
