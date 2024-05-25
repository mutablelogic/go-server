package logger

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/router"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	remoteAddrHeaders = []string{"X-Real-Ip", "X-Forwarded-For"}
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (l *logger) Wrap(ctx context.Context, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nw := NewResponseWriter(w)
		next(nw, r)

		// Print the response
		result := fmt.Sprintf("%v %v %v -> [%v]", remoteAddr(r), r.Method, r.URL, nw.Status())
		if sz := nw.Size(); sz > 0 {
			result += fmt.Sprintf(" %v bytes written", sz)
		}
		if t := router.Time(r.Context()); !t.IsZero() {
			result += fmt.Sprint(" in ", time.Since(t).Truncate(time.Microsecond))
		}
		l.Printf(r.Context(), result)
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
