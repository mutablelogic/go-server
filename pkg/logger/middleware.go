package logger

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	remoteAddrHeaders = []string{"X-Real-Ip", "X-Forwarded-For", "CF-Connecting-IP", "True-Client-IP"}
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

// NewMiddleware returns an HTTP middleware function that logs each request.
// If log is nil, slog.Default() is used.
func NewMiddleware(log *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			nw := NewResponseWriter(w)
			next(nw, r)

			// Print the response
			result := fmt.Sprintf("%v %v -> [%v]", r.Method, r.URL, nw.Status())
			log.With(
				"delta_ms", time.Since(now).Milliseconds(),
				"method", r.Method,
				"status", nw.Status(),
				"remote_ip", remoteAddr(r),
				"url", r.URL.String(),
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"host", r.URL.Host,
				"size", nw.Size(),
			).InfoContext(r.Context(), result)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

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
