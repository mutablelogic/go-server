package httprouter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

func Test_Middleware_001(t *testing.T) {
	assert := assert.New(t)

	// Wrap with no middleware should return the handler unchanged
	var mw middlewareFuncs
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}

	wrapped := mw.Wrap(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("ok", rec.Body.String())
}

func Test_Middleware_002(t *testing.T) {
	assert := assert.New(t)

	// Wrap with a single middleware that adds a header
	mw := middlewareFuncs{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Single", "applied")
				next(w, r)
			}
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrapped := mw.Wrap(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal("applied", rec.Header().Get("X-Single"))
}

func Test_Middleware_003(t *testing.T) {
	assert := assert.New(t)

	// Wrap with multiple middleware, verify execution order.
	// Wrap iterates the slice in reverse, so the first element ends up as the
	// outermost wrapper and executes first when the request arrives.
	var order []string

	first := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "first")
			next(w, r)
		}
	}
	second := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "second")
			next(w, r)
		}
	}

	mw := middlewareFuncs{first, second}

	handler := func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	}

	wrapped := mw.Wrap(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal([]string{"first", "second", "handler"}, order)
}

func Test_Middleware_004(t *testing.T) {
	assert := assert.New(t)

	// Middleware can short-circuit the chain by not calling next
	mw := middlewareFuncs{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("denied"))
				// next is intentionally not called
			}
		},
	}

	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	wrapped := mw.Wrap(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(rec, req)

	assert.Equal(http.StatusForbidden, rec.Code)
	assert.Equal("denied", rec.Body.String())
	assert.False(called)
}
