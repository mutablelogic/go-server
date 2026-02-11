package logger

import (
	"fmt"
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

func (t *Logger) WrapFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		nw := NewResponseWriter(w)
		next(nw, r)

		// Print the response
		result := fmt.Sprintf("%v %v -> [%v]", r.Method, r.URL, nw.Status())
		t.With(
			"delta_ms", time.Since(now).Milliseconds(),
			"method", r.Method,
			"status", nw.Status(),
			"remote_ip", remoteAddr(r),
			"url", r.URL.String(),
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"host", r.URL.Host,
			"size", nw.Size(),
		).Print(r.Context(), result)
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
