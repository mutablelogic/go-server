package otel_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"mime/multipart"
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
	assert.NotContains(t, spans[0].Name, token)
	assert.Contains(t, spans[0].Name, "/hook/{token}")

	// A path value which isn't marked is logged, as it is
	logs.Reset()
	exporter.Reset()
	req = httptest.NewRequest(http.MethodGet, "/item/42", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	provider.ForceFlush(req.Context())
	assert.Contains(t, logs.String(), `msg="request GET /item/42 -> 200 OK"`)
	require.Len(t, exporter.GetSpans(), 1)
	assert.Contains(t, exporter.GetSpans()[0].Attributes, attribute.String("url.path", "/item/42"))
	assert.Contains(t, exporter.GetSpans()[0].Name, "/item/{id}", "the span is named by the mux's pattern")

	// A request which the middleware isn't serving is ignored
	plain := httptest.NewRequest(http.MethodPost, "/hook/x", nil)
	assert.NotPanics(t, func() { otel.RedactPathValue(plain, "token") })
}

func TestHTTPHandler_RedactPathValueCases(t *testing.T) {
	for _, test := range []struct {
		name, pattern, target string
		names                 []string
		path, full            string
		secrets               []string
	}{
		{
			name: "Escaped", pattern: "/hook/{token}", target: "/hook/%73ecretTOKENvalue?x=1", names: []string{"token"},
			path: "/hook/{token}", full: "/hook/{token}?x=1", secrets: []string{"ecretTOKENvalue"},
		},
		{
			name: "Escaped slash", pattern: "/hook/{token}/a%20b", target: "/hook/sec%2Fret/a%20b", names: []string{"token"},
			path: "/hook/{token}/a b", full: "/hook/{token}/a%20b", secrets: []string{"ret/", "%2Fret"},
		},
		{
			name: "Overlapping", pattern: "/hook/{short}/{long}", target: "/hook/abc/abcdef123456", names: []string{"short", "long"},
			path: "/hook/{short}/{long}", full: "/hook/{short}/{long}", secrets: []string{"abc", "def123456"},
		},
		{
			name: "More than one segment", pattern: "/hook/{rest...}", target: "/hook/one/two/three", names: []string{"rest"},
			path: "/hook/{rest}", full: "/hook/{rest}", secrets: []string{"one", "two", "three"},
		},
		{
			name: "Within another segment", pattern: "/hook/{token}", target: "/hook/ook", names: []string{"token"},
			path: "/hook/{token}", full: "/hook/{token}",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			exporter, provider := newTestTracer()
			setGlobalTracerProvider(t, provider)
			mux := http.NewServeMux()
			called := false
			mux.HandleFunc(test.pattern, func(w http.ResponseWriter, r *http.Request) {
				called = true
				otel.RedactPathValue(r, test.names...)
			})
			var logs bytes.Buffer
			handler := otel.HTTPHandler("test", slog.New(slog.NewJSONHandler(&logs, nil)))(mux)
			req := httptest.NewRequest(http.MethodGet, test.target, nil)
			handler.ServeHTTP(httptest.NewRecorder(), req)
			provider.ForceFlush(req.Context())
			require.True(t, called)

			var logged map[string]any
			require.NoError(t, json.Unmarshal(logs.Bytes(), &logged))
			assert.Equal(t, test.path, logged["url.path"])
			assert.Equal(t, test.full, logged["url.full"])
			spans := exporter.GetSpans()
			require.Len(t, spans, 1)
			assert.Contains(t, spans[0].Attributes, attribute.String("url.path", test.path))
			for _, secret := range test.secrets {
				assert.NotContains(t, logs.String(), secret)
				assert.NotContains(t, spans[0].Name, secret)
				for _, attr := range spans[0].Attributes {
					assert.NotContains(t, attr.Value.Emit(), secret, string(attr.Key))
				}
			}
		})
	}
}

func TestHTTPHandler_RedactPathValueWithoutPattern(t *testing.T) {
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	// A router whose pattern otelhttp doesn't see, as it serves a copy of the
	// request, whose span would be named by its path
	mux := http.NewServeMux()
	mux.HandleFunc("/hook/{token}", func(w http.ResponseWriter, r *http.Request) {
		otel.RedactPathValue(r, "token")
	})
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r.WithContext(r.Context()))
	})
	req := httptest.NewRequest(http.MethodPost, "/hook/s3cretTOKENvalue", nil)
	otel.HTTPHandler("test", newTestLogger())(router).ServeHTTP(httptest.NewRecorder(), req)
	provider.ForceFlush(req.Context())
	require.Len(t, exporter.GetSpans(), 1)
	assert.Equal(t, "POST /hook/{token}", exporter.GetSpans()[0].Name)
}

func TestHTTPHandler_RedactPathValuePanic(t *testing.T) {
	exporter, provider := newTestTracer()
	setGlobalTracerProvider(t, provider)

	// A handler which panics, whose panic isn't swallowed
	mux := http.NewServeMux()
	mux.HandleFunc("/hook/{token}", func(w http.ResponseWriter, r *http.Request) {
		otel.RedactPathValue(r, "token")
		panic("failed")
	})
	const token = "s3cretTOKENvalue"
	req := httptest.NewRequest(http.MethodPost, "/hook/"+token, nil)
	assert.PanicsWithValue(t, "failed", func() {
		otel.HTTPHandler("test", newTestLogger())(mux).ServeHTTP(httptest.NewRecorder(), req)
	})
	provider.ForceFlush(req.Context())
	require.Len(t, exporter.GetSpans(), 1)
	span := exporter.GetSpans()[0]
	assert.NotContains(t, span.Name, token)
	assert.Contains(t, span.Attributes, attribute.String("url.path", "/hook/{token}"))
	for _, attr := range span.Attributes {
		assert.NotContains(t, attr.Value.Emit(), token, string(attr.Key))
	}
}

func TestHTTPHandler_MultipartForm(t *testing.T) {
	// The form which the handler parses is the request's which the server
	// made, so it removes its files
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	require.NoError(t, form.WriteField("a", "1"))
	require.NoError(t, form.Close())
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	otel.HTTPHandler("test", newTestLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, r.ParseMultipartForm(1<<20))
	})).ServeHTTP(httptest.NewRecorder(), req)
	require.NotNil(t, req.MultipartForm)
	assert.Equal(t, []string{"1"}, req.MultipartForm.Value["a"])
}
