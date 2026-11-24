package httprouter_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	otel "github.com/mutablelogic/go-server/pkg/otel"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	gootel "go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func Test_Router_SecretPathParameter(t *testing.T) {
	const secret = "s3cretTOKENvalue"
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	gootel.SetTracerProvider(provider)
	t.Cleanup(func() { gootel.SetTracerProvider(tracenoop.NewTracerProvider()) })

	// A middleware, which refuses some requests before the handler
	refuse := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Refuse") != "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	var logs bytes.Buffer
	router, err := httprouter.NewRouter(context.Background(), http.NewServeMux(), "/api", "*", "test", "v1",
		otel.HTTPHandlerFunc("test", slog.New(slog.NewTextHandler(&logs, nil))), refuse)
	require.NoError(t, err)

	// A route whose token is secret, whose handlers don't mark it, and one
	// whose id isn't
	type hookParams struct {
		Token string `json:"token" format:"password"`
	}
	type itemParams struct {
		ID string `json:"id" format:"uuid"`
	}
	var token string
	require.NoError(t, router.Register("hook/{token}", jsonschema.MustFor[hookParams](), func(path httprequest.PathItem) {
		path.Post(func(w http.ResponseWriter, r *http.Request) {
			token = r.PathValue("token")
			w.WriteHeader(http.StatusAccepted)
		}, func(op httprequest.PathOperation) {})
		path.Put(func(http.ResponseWriter, *http.Request) { panic("failed") }, func(op httprequest.PathOperation) {})
	}))
	require.NoError(t, router.Register("item/{id}", jsonschema.MustFor[itemParams](), func(path httprequest.PathItem) {
		path.Get(func(w http.ResponseWriter, r *http.Request) {}, func(op httprequest.PathOperation) {})
	}))
	do := func(method, target string, header ...string) int {
		logs.Reset()
		exporter.Reset()
		w, r := httptest.NewRecorder(), httptest.NewRequest(method, target, nil)
		for i := 0; i+1 < len(header); i += 2 {
			r.Header.Set(header[i], header[i+1])
		}
		router.ServeHTTP(w, r)
		return w.Code
	}
	check := func(t *testing.T) {
		t.Helper()
		assert.NotContains(t, logs.String(), secret)
		require.Len(t, exporter.GetSpans(), 1)
		assert.NotContains(t, exporter.GetSpans()[0].Name, secret)
		for _, attr := range exporter.GetSpans()[0].Attributes {
			assert.NotContains(t, attr.Value.Emit(), secret, string(attr.Key))
		}
	}

	t.Run("Handled", func(t *testing.T) {
		assert.Equal(t, http.StatusAccepted, do(http.MethodPost, "/api/hook/"+secret))
		assert.Equal(t, secret, token, "the handler still has it")
		check(t)
		assert.Contains(t, logs.String(), "request POST /api/hook/{token} -> 202")
	})
	t.Run("Method not allowed", func(t *testing.T) {
		assert.Equal(t, http.StatusMethodNotAllowed, do(http.MethodDelete, "/api/hook/"+secret))
		check(t)
		assert.Contains(t, logs.String(), "request DELETE /api/hook/{token} -> 405")
	})
	t.Run("Refused by a middleware", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, do(http.MethodPost, "/api/hook/"+secret, "X-Refuse", "1"))
		check(t)
	})
	t.Run("Panics", func(t *testing.T) {
		assert.PanicsWithValue(t, "failed", func() { do(http.MethodPut, "/api/hook/"+secret) })
		check(t)
	})
	t.Run("Not secret", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, do(http.MethodGet, "/api/item/42"))
		assert.Contains(t, logs.String(), "request GET /api/item/42 -> 200")
	})
}
