package otel

import (
	"net/http"

	// Packages
	otel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	propagation "go.opentelemetry.io/otel/propagation"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// HTTP SERVER MIDDLEWARE

// HTTPHandler returns an HTTP middleware that creates OpenTelemetry spans for
// incoming requests. It extracts trace context from request headers to support
// distributed tracing, records the HTTP method, URL, and status code, and marks
// spans as errors for 4xx/5xx responses.
func HTTPHandler(tracer trace.Tracer) func(http.Handler) http.Handler {
	wrap := HTTPHandlerFunc(tracer)
	return func(next http.Handler) http.Handler {
		return wrap(next.ServeHTTP)
	}
}

// HTTPHandlerFunc is like HTTPHandler but wraps http.HandlerFunc directly.
// This is useful for routers that use HandlerFunc.
func HTTPHandlerFunc(tracer trace.Tracer) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from incoming request headers
			parentCtx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			ctx, span := tracer.Start(parentCtx, r.Method+" "+r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.request.method", r.Method),
					attribute.String("url.full", r.URL.String()),
					attribute.String("url.path", r.URL.Path),
					attribute.String("server.address", r.Host),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			// Record status code and mark errors
			span.SetAttributes(attribute.Int("http.response.status_code", sw.status))
			if sw.status >= 400 {
				span.SetStatus(codes.Error, http.StatusText(sw.status))
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE TYPES

// statusWriter wraps http.ResponseWriter to capture the status code.
// It implements [http.Flusher] and Unwrap so that downstream middleware
// (such as the logger) and streaming responses (SSE) work correctly.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap returns the underlying [http.ResponseWriter], allowing
// [http.ResponseController] and type assertions to reach it.
func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Flush sends any buffered data to the client. It delegates to the
// underlying writer when it implements [http.Flusher].
func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
