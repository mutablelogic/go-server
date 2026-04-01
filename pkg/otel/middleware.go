package otel

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	// Packages

	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	attribute "go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"
)

var remoteAddrHeaders = []string{"X-Real-Ip", "X-Forwarded-For", "CF-Connecting-IP", "True-Client-IP"}

///////////////////////////////////////////////////////////////////////////////
// HTTP SERVER MIDDLEWARE

// HTTPHandler returns an HTTP middleware that instruments requests with
// otelhttp and logs a request summary after the handler completes.
func HTTPHandler(serverName string, log *slog.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped := NewResponseWriter(w)
			start := time.Now()

			next.ServeHTTP(wrapped, r)

			if r.Pattern != "" {
				if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
					labeler.Add(attribute.String("http.route", r.Pattern))
				}
			}

			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}
			statusText := http.StatusText(status)

			attrs := []any{
				"event.duration", time.Since(start).Nanoseconds(),
				"http.request.method", r.Method,
				"http.response.status_code", status,
				"http.response.status_text", statusText,
				"http.response.body.size", wrapped.Size(),
				"client.address", remoteAddr(r),
				"url.full", r.URL.String(),
				"url.path", r.URL.Path,
				"url.query", r.URL.RawQuery,
				"server.address", r.Host,
				"server.name", serverName,
			}

			if spanCtx := trace.SpanFromContext(r.Context()).SpanContext(); spanCtx.IsValid() {
				attrs = append(attrs, "trace_id", spanCtx.TraceID().String())
			}

			msg := fmt.Sprintf("request %s %s -> %d", r.Method, r.URL.Path, status)
			if statusText != "" {
				msg += " " + statusText
			}

			switch {
			case status >= 100 && status < 400:
				log.InfoContext(r.Context(), msg, attrs...)
			case status >= 400 && status < 500:
				log.WarnContext(r.Context(), msg, attrs...)
			case status >= 500 && status < 600:
				log.ErrorContext(r.Context(), msg, attrs...)
			default:
				log.DebugContext(r.Context(), msg, attrs...)
			}
		})

		return otelhttp.NewHandler(
			handler,
			serverName,
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				route := r.Pattern
				if route == "" {
					route = r.URL.Path
				}
				return r.Method + " " + route
			}),
		)
	}
}

// HTTPHandlerFunc is like HTTPHandler but wraps http.HandlerFunc directly.
// This is useful for routers that use HandlerFunc.
func HTTPHandlerFunc(serverAddress string, log *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	wrap := HTTPHandler(serverAddress, log)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return wrap(next).ServeHTTP
	}
}

func remoteAddr(r *http.Request) string {
	for _, header := range remoteAddrHeaders {
		if addr := r.Header.Get(header); addr != "" {
			return addr
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
