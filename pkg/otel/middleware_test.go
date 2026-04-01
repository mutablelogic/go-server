package otel_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	otel "github.com/mutablelogic/go-server/pkg/otel"
	assert "github.com/stretchr/testify/assert"
	gootel "go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

///////////////////////////////////////////////////////////////////////////////
// TEST HELPERS

func newTestTracer() (*tracetest.InMemoryExporter, *sdktrace.TracerProvider) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	return exporter, provider
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func setGlobalTracerProvider(t *testing.T, provider *sdktrace.TracerProvider) {
	t.Helper()
	gootel.SetTracerProvider(provider)
	t.Cleanup(func() {
		gootel.SetTracerProvider(tracenoop.NewTracerProvider())
	})
}

func TestHTTPHandler_CreatesSpan(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	handler := otel.HTTPHandler("test", newTestLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusOK, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Equal("GET /api/test", spans[0].Name)
}

func TestHTTPHandlerFunc_CreatesSpan(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	handler := otel.HTTPHandlerFunc("test", newTestLogger())(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/create", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusCreated, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Equal("POST /api/create", spans[0].Name)
}

func TestHTTPHandler_CapturesStatusCode(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	handler := otel.HTTPHandler("test", newTestLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusNotFound, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}

func TestHTTPHandler_DefaultStatusOK(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	// Handler that doesn't explicitly call WriteHeader
	handler := otel.HTTPHandler("test", newTestLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("implicit 200"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusOK, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}

func TestHTTPHandler_ServerError(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	handler := otel.HTTPHandler("test", newTestLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusInternalServerError, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}
