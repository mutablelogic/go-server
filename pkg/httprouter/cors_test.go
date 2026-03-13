package httprouter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

func Test_Cors_001(t *testing.T) {
	assert := assert.New(t)

	// Default wildcard origin (empty string)
	mw := Cors("")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func Test_Cors_002(t *testing.T) {
	assert := assert.New(t)

	// Specific origin
	mw := Cors("https://example.com")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func Test_Cors_003(t *testing.T) {
	assert := assert.New(t)

	// Preflight request without headers list defaults to "*"
	mw := Cors("https://example.com")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		// Should not be called for OPTIONS
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Equal("https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("*", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal("*", rec.Header().Get("Access-Control-Allow-Headers"))
}

func Test_Cors_004(t *testing.T) {
	assert := assert.New(t)

	// Preflight with explicit allowed headers
	mw := Cors("*", "Accept", "Content-Type", "Authorization")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Equal("*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal("Accept, Content-Type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
}

func Test_Cors_005(t *testing.T) {
	assert := assert.New(t)

	// Non-preflight requests must pass through to the next handler
	called := false
	mw := Cors("*", "Content-Type")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.True(called)
	assert.Equal(http.StatusCreated, rec.Code)
	assert.Equal("*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func Test_Cors_006(t *testing.T) {
	assert := assert.New(t)

	// Invalid header keys are silently filtered; valid ones remain
	mw := Cors("*", "Content-Type", "invalid header", "Authorization")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Equal("Content-Type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
}

func Test_Cors_007(t *testing.T) {
	assert := assert.New(t)

	// All headers invalid → falls back to "*"
	mw := Cors("*", "bad header", "also bad!")
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(http.StatusNoContent, rec.Code)
	assert.Equal("*", rec.Header().Get("Access-Control-Allow-Headers"))
}
