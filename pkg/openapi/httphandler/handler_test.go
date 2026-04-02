package httphandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	httphandler "github.com/mutablelogic/go-server/pkg/openapi/httphandler"
	yaml "gopkg.in/yaml.v3"
)

func newRouter(t *testing.T) *httprouter.Router {
	t.Helper()
	router, err := httprouter.NewRouter(context.Background(), http.NewServeMux(), "/api", "", "Test API", "v1")
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return router
}

func Test_RegisterHandler_JSON(t *testing.T) {
	router := newRouter(t)
	if err := httphandler.RegisterHandler(router); err != nil {
		t.Fatalf("RegisterHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET openapi.json status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
	var spec map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if spec["openapi"] == nil {
		t.Error("JSON spec missing 'openapi' field")
	}
}

func Test_RegisterHandler_YAML(t *testing.T) {
	router := newRouter(t)
	if err := httphandler.RegisterHandler(router); err != nil {
		t.Fatalf("RegisterHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET openapi.yaml status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/yaml" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/yaml")
	}
	var spec map[string]any
	if err := yaml.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("response is not valid YAML: %v", err)
	}
	if spec["openapi"] == nil {
		t.Error("YAML spec missing 'openapi' field")
	}
}

func Test_RegisterHandler_MethodNotAllowed(t *testing.T) {
	router := newRouter(t)
	if err := httphandler.RegisterHandler(router); err != nil {
		t.Fatalf("RegisterHandler: %v", err)
	}

	for _, path := range []string{"/api/openapi.json", "/api/openapi.yaml", "/api/openapi.html"} {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			req := httptest.NewRequest(method, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s = %d, want %d", method, path, rec.Code, http.StatusMethodNotAllowed)
			}
		}
	}
}

func Test_RegisterHandler_Duplicate(t *testing.T) {
	router := newRouter(t)
	if err := httphandler.RegisterHandler(router); err != nil {
		t.Fatalf("first RegisterHandler: %v", err)
	}
	if err := httphandler.RegisterHandler(router); err == nil {
		t.Error("second RegisterHandler should return an error for duplicate paths, got nil")
	}
}

func Test_RegisterHandler_HTML(t *testing.T) {
	router := newRouter(t)
	if err := httphandler.RegisterHandler(router); err != nil {
		t.Fatalf("RegisterHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/openapi.html", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET openapi.html status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("response body is empty")
	}
	if !strings.Contains(body, "redoc") {
		t.Error("response body does not contain expected content")
	}
}
