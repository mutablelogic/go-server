package httprouter

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	// Packages
	"github.com/stretchr/testify/assert"
)

func Test_Router_001(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with no origin
	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0")
	assert.NoError(err)
	assert.NotNil(router)
}

func Test_Router_002(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with wildcard origin
	router, err := NewRouter(context.Background(), "/", "*", "Test API", "1.0.0")
	assert.NoError(err)
	assert.NotNil(router)
	assert.Equal("*", router.Origin())
}

func Test_Router_003(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with a specific origin
	router, err := NewRouter(context.Background(), "/", "https://example.com", "Test API", "1.0.0")
	assert.NoError(err)
	assert.NotNil(router)
	assert.Equal("https://example.com", router.Origin())
}

func Test_Router_004(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with an invalid origin
	_, err := NewRouter(context.Background(), "/", "not-a-valid-origin", "Test API", "1.0.0")
	assert.Error(err)
}

func Test_RegisterNotFound_001(t *testing.T) {
	assert := assert.New(t)

	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0")
	assert.NoError(err)
	assert.NoError(router.RegisterNotFound("/", false))

	// Request a path that doesn't match any specific handler
	req := httptest.NewRequest(http.MethodGet, "/unknown-path", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusNotFound, rec.Code)
}

func Test_RegisterOpenAPI_001(t *testing.T) {
	assert := assert.New(t)

	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0")
	assert.NoError(err)
	assert.NoError(router.RegisterOpenAPI("/api/v1", false))

	// GET should return the OpenAPI spec as JSON
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Contains(rec.Header().Get("Content-Type"), "application/json")

	// Decode the response body and check the spec
	var spec map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &spec)
	assert.NoError(err)
	assert.Equal("3.1.1", spec["openapi"])

	info, ok := spec["info"].(map[string]any)
	assert.True(ok)
	assert.Equal("Test API", info["title"])
	assert.Equal("1.0.0", info["version"])
}

func Test_RegisterOpenAPI_002(t *testing.T) {
	assert := assert.New(t)

	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0")
	assert.NoError(err)
	assert.NoError(router.RegisterOpenAPI("/api/v1", false))

	// POST should return method not allowed
	req := httptest.NewRequest(http.MethodPost, "/api/v1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusMethodNotAllowed, rec.Code)
}

func Test_RegisterFS_001(t *testing.T) {
	assert := assert.New(t)

	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0")
	assert.NoError(err)

	// Create an in-memory filesystem with a test file
	fsys := fstest.MapFS{
		"test.txt": &fstest.MapFile{
			Data: []byte("Hello"),
		},
	}
	assert.NoError(router.RegisterFS("/", fs.FS(fsys), false, nil))

	// Request the file from the root-served filesystem
	req := httptest.NewRequest(http.MethodGet, "/test.txt", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Contains(rec.Body.String(), "Hello")
}

func Test_RegisterFunc_001(t *testing.T) {
	assert := assert.New(t)

	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0")
	assert.NoError(err)

	// Register a custom handler
	assert.NoError(router.RegisterFunc("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}), false, nil))

	// Request the handler
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("hello world", rec.Body.String())
}

func Test_RegisterFunc_002(t *testing.T) {
	assert := assert.New(t)

	// Create a router with middleware that adds a custom header
	mw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next(w, r)
		}
	}

	router, err := NewRouter(context.Background(), "/", "", "Test API", "1.0.0", mw)
	assert.NoError(err)

	// Register a handler with middleware enabled
	assert.NoError(router.RegisterFunc("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), true, nil))

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("middleware", rec.Header().Get("X-Test"))
}

func Test_RegisterFunc_003(t *testing.T) {
	assert := assert.New(t)

	// Create a router with a prefix
	router, err := NewRouter(context.Background(), "/api/v1", "", "Test API", "1.0.0")
	assert.NoError(err)

	// Register a handler under the prefix
	assert.NoError(router.RegisterFunc("/items", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("items"))
	}), false, nil))

	// Request using the full prefixed path
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("items", rec.Body.String())
}
