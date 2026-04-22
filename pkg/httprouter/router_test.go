package httprouter

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi_op "github.com/mutablelogic/go-server/pkg/openapi"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	assert "github.com/stretchr/testify/assert"
)

func newTestRouter(t *testing.T, prefix, origin string, middleware ...HTTPMiddlewareFunc) *Router {
	t.Helper()
	router, err := NewRouter(context.Background(), http.NewServeMux(), prefix, origin, "Test API", "1.0.0", middleware...)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return router
}

func Test_Router_001(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with no origin
	router := newTestRouter(t, "/", "")
	assert.NotNil(router)
}

func Test_Router_002(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with wildcard origin
	router := newTestRouter(t, "/", "*")
	assert.NotNil(router)
	assert.Equal("*", router.Origin())
}

func Test_Router_003(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with a specific origin
	router := newTestRouter(t, "/", "https://example.com")
	assert.NotNil(router)
	assert.Equal("https://example.com", router.Origin())
}

func Test_Router_004(t *testing.T) {
	assert := assert.New(t)

	// Create a new router with an invalid origin
	_, err := NewRouter(context.Background(), http.NewServeMux(), "/", "not-a-valid-origin", "Test API", "1.0.0")
	assert.Error(err)
}

func Test_Router_005(t *testing.T) {
	assert := assert.New(t)

	// A nil ServeMux should fail cleanly instead of panicking later.
	_, err := NewRouter(context.Background(), nil, "/", "", "Test API", "1.0.0")
	assert.Error(err)
	assert.Contains(err.Error(), "mux is nil")
}

func Test_RegisterCatchAll_001(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")
	assert.NoError(router.RegisterCatchAll("/", false))

	// Request a path that doesn't match any specific handler
	req := httptest.NewRequest(http.MethodGet, "/unknown-path", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusNotFound, rec.Code)
}

func Test_RegisterOpenAPI_001(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")
	assert.NoError(router.RegisterOpenAPI("/api/v1", false))

	// GET should return the OpenAPI spec as JSON
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Contains(rec.Header().Get("Content-Type"), "application/json")

	// Decode the response body and check the spec
	var spec map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &spec)
	assert.NoError(err)
	assert.Equal("3.1.1", spec["openapi"])

	info, ok := spec["info"].(map[string]any)
	assert.True(ok)
	assert.Equal("Test API", info["title"])
	assert.Equal("1.0.0", info["version"])
}

func Test_RegisterOpenAPI_002(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")
	assert.NoError(router.RegisterOpenAPI("/api/v1", false))

	// POST should return method not allowed
	req := httptest.NewRequest(http.MethodPost, "/api/v1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusMethodNotAllowed, rec.Code)
}

func Test_RegisterFS_001(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")

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

func Test_RegisterFS_002(t *testing.T) {
	assert := assert.New(t)
	modified := time.Date(2026, time.April, 22, 12, 0, 0, 0, time.UTC)
	handler := withLastModified(modified, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello"))
	}))

	condReq := httptest.NewRequest(http.MethodGet, "/test.txt", nil)
	condReq.Header.Set("If-Modified-Since", modified.Format(http.TimeFormat))
	condRec := httptest.NewRecorder()
	handler.ServeHTTP(condRec, condReq)

	assert.Equal(http.StatusNotModified, condRec.Code)
	assert.Empty(condRec.Body.String())
	assert.Equal(modified.Format(http.TimeFormat), condRec.Header().Get("Last-Modified"))
}

func Test_RegisterPath_UsingGet_001(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")

	item := httprequest.NewPathItem("Hello", "Hello route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	}, "Get hello")
	assert.NoError(router.RegisterPath("/hello", nil, item))

	// Request the handler
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("hello world", rec.Body.String())
}

func Test_RegisterPath_UsingGet_002(t *testing.T) {
	assert := assert.New(t)

	// Create a router with middleware that adds a custom header
	mw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next(w, r)
		}
	}

	router := newTestRouter(t, "/", "", mw)

	item := httprequest.NewPathItem("Hello", "Hello route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "Get hello")
	assert.NoError(router.RegisterPath("/hello", nil, item))

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("middleware", rec.Header().Get("X-Test"))
}

func Test_RegisterPath_UsingGet_003(t *testing.T) {
	assert := assert.New(t)

	// Create a router with a prefix
	router := newTestRouter(t, "/api/v1", "")

	// Register a relative path - should be joined with the prefix
	item := httprequest.NewPathItem("Items", "Items route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("items"))
	}, "Get items")
	assert.NoError(router.RegisterPath("items", nil, item))

	// Request using the full prefixed path
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("items", rec.Body.String())
}

func Test_RegisterPath_UsingGet_004(t *testing.T) {
	assert := assert.New(t)

	// Create a router with a prefix
	router := newTestRouter(t, "/api/v1", "")

	// Register an absolute path - prefix should not be prepended
	item := httprequest.NewPathItem("Health", "Health route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}, "Get health")
	assert.NoError(router.RegisterPath("/health", nil, item))

	// Request at the absolute path, not under the prefix
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("ok", rec.Body.String())
}

// mockPathItem is a minimal implementation of httprequest.PathItem for testing.
type mockPathItem struct {
	handler  http.HandlerFunc
	handlers map[string]http.HandlerFunc
	spec     *openapi.PathItem
}

type testSecurityScheme struct{}

func (m *mockPathItem) Handler() http.HandlerFunc {
	if len(m.handlers) == 0 {
		return m.handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := m.handlers[r.Method]; ok && handler != nil {
			handler(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *mockPathItem) Spec(path string, params *jsonschema.Schema) *openapi.PathItem {
	return m.spec
}

func (m *mockPathItem) WrapHandler(method string, fn func(http.HandlerFunc) http.HandlerFunc) {
	if h, ok := m.handlers[method]; ok && h != nil {
		m.handlers[method] = fn(h)
	}
}

func (testSecurityScheme) Wrap(handler http.HandlerFunc, scopes []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Security-Scopes", strings.Join(scopes, ","))
		handler(w, r)
	}
}

func (testSecurityScheme) Spec() openapi.SecurityScheme {
	return openapi.SecurityScheme{Type: "http", Scheme: "bearer"}
}

func Test_RegisterPath_001(t *testing.T) {
	assert := assert.New(t)

	// RegisterPath with a relative path joins the prefix
	router := newTestRouter(t, "/api/v1", "")

	called := false
	item := &mockPathItem{
		handler: func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		},
		spec: &openapi.PathItem{Summary: "Test"},
	}
	assert.NoError(router.RegisterPath("items", nil, item))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.True(called)
	assert.Equal(http.StatusOK, rec.Code)
}

func Test_RegisterPath_002(t *testing.T) {
	assert := assert.New(t)

	// RegisterPath with an absolute path ignores the prefix
	router := newTestRouter(t, "/api/v1", "")

	item := &mockPathItem{
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("absolute"))
		},
	}
	assert.NoError(router.RegisterPath("/health", nil, item))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("absolute", rec.Body.String())
}

func Test_RegisterPath_003(t *testing.T) {
	assert := assert.New(t)

	// RegisterPath adds spec to the OpenAPI document
	router := newTestRouter(t, "/api", "")

	item := &mockPathItem{
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		spec: &openapi.PathItem{
			Summary: "List widgets",
			Get:     &openapi.Operation{Summary: "Get widgets"},
		},
	}
	assert.NoError(router.RegisterPath("widgets", nil, item))

	spec := router.Spec()
	assert.NotNil(spec)

	// Marshal and check the spec contains the path
	data, err := json.Marshal(spec)
	assert.NoError(err)

	var raw map[string]any
	assert.NoError(json.Unmarshal(data, &raw))
	paths, ok := raw["paths"].(map[string]any)
	assert.True(ok)
	_, ok = paths["/api/widgets"]
	assert.True(ok, "expected /api/widgets in spec paths")
}

func Test_RegisterPath_004(t *testing.T) {
	assert := assert.New(t)

	// RegisterPath with nil handler returns an error
	router := newTestRouter(t, "/", "")

	item := &mockPathItem{handler: nil}
	err := router.RegisterPath("broken", nil, item)
	assert.Error(err)
}

func Test_RegisterPath_005(t *testing.T) {
	assert := assert.New(t)

	// RegisterPath always applies middleware
	mw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next(w, r)
		}
	}

	router := newTestRouter(t, "/", "", mw)

	item := &mockPathItem{
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}
	assert.NoError(router.RegisterPath("test", nil, item))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("middleware", rec.Header().Get("X-Test"))
}

func Test_RegisterPath_006(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")
	assert.NoError(router.RegisterSecurityScheme("bearerAuth", testSecurityScheme{}))

	item := httprequest.NewPathItem("Secure", "Secure route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, "Get secure route", openapi_op.WithSecurity("bearerAuth", "read"))

	assert.NoError(router.RegisterPath("secure", nil, item))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Equal("read", rec.Header().Get("X-Security-Scopes"))
}

func Test_RegisterPath_007(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")

	item := httprequest.NewPathItem("Secure", "Secure route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, "Get secure route", openapi_op.WithSecurity("missingAuth", "read"))

	err := router.RegisterPath("secure", nil, item)
	assert.Error(err)
	assert.Contains(err.Error(), "security scheme \"missingAuth\" not registered")
}

func Test_RegisterPath_008(t *testing.T) {
	assert := assert.New(t)

	router := newTestRouter(t, "/", "")

	item := httprequest.NewPathItem("Secure", "Secure route")
	item.Get(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, "Get secure route", openapi_op.WithSecurity("missingAuth", "read"))

	err := router.RegisterPath("secure", nil, item)
	assert.Error(err)

	spec := router.Spec()
	if assert.NotNil(spec) {
		if spec.Paths != nil {
			_, exists := spec.Paths.MapOfPathItemValues["/secure"]
			assert.False(exists)
		}
	}
}

func Test_mockPathItem_WrapHandler_001(t *testing.T) {
	assert := assert.New(t)

	called := false
	item := &mockPathItem{
		handlers: map[string]http.HandlerFunc{
			http.MethodGet: func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			},
		},
	}

	item.WrapHandler(http.MethodPost, func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Wrapped", "true")
			next(w, r)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	item.Handler()(rec, req)

	assert.True(called)
	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Empty(rec.Header().Get("X-Wrapped"))

	called = false
	item.WrapHandler(http.MethodGet, func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Wrapped", "true")
			next(w, r)
		}
	})

	req = httptest.NewRequest(http.MethodGet, "/items", nil)
	rec = httptest.NewRecorder()
	item.Handler()(rec, req)

	assert.True(called)
	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Equal("true", rec.Header().Get("X-Wrapped"))
}
