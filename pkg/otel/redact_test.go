package otel_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	otel "github.com/mutablelogic/go-server/pkg/otel"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	attribute "go.opentelemetry.io/otel/attribute"
)

func TestHTTPHandler_RedactPathValue(t *testing.T) {
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	// A route whose token is secret, and one whose id isn't
	mux := http.NewServeMux()
	mux.HandleFunc("POST /hook/{token}", func(w http.ResponseWriter, r *http.Request) {
		otel.RedactPathValue(r, "token", "missing")
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /item/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	var logs bytes.Buffer
	handler := otel.HTTPHandler("test", slog.New(slog.NewTextHandler(&logs, nil)))(mux)

	// The token isn't in the log, nor the span, and its name is in its place
	const token = "s3cretTOKENvalue"
	req := httptest.NewRequest(http.MethodPost, "/hook/"+token+"?x=1", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	provider.ForceFlush(req.Context())
	logged := logs.String()
	assert.NotContains(t, logged, token)
	assert.Contains(t, logged, `msg="request POST /hook/{token} -> 202 Accepted"`)
	assert.Contains(t, logged, "url.path=/hook/{token}")
	assert.Contains(t, logged, "url.full=\"/hook/{token}?x=1\"")
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	for _, attr := range spans[0].Attributes {
		assert.NotContains(t, attr.Value.Emit(), token, string(attr.Key))
	}
	assert.Contains(t, spans[0].Attributes, attribute.String("url.path", "/hook/{token}"))

	// A path value which isn't marked is logged, as it is
	logs.Reset()
	exporter.Reset()
	req = httptest.NewRequest(http.MethodGet, "/item/42", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	provider.ForceFlush(req.Context())
	assert.Contains(t, logs.String(), `msg="request GET /item/42 -> 200 OK"`)
	require.Len(t, exporter.GetSpans(), 1)
	assert.Contains(t, exporter.GetSpans()[0].Attributes, attribute.String("url.path", "/item/42"))

	// A request which the middleware isn't serving is ignored
	plain := httptest.NewRequest(http.MethodPost, "/hook/x", nil)
	assert.NotPanics(t, func() { otel.RedactPathValue(plain, "token") })
}
