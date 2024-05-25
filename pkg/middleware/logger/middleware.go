package logger

import (
	"context"
	"net/http"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/router"
)

func (l *logger) Wrap(ctx context.Context, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nw := NewResponseWriter(w)

		// Call the next handler
		next(nw, r)

		// Print the response
		l.Printf(r.Context(), "%v %v -> [%v] %q", r.Method, r.URL, nw.Status(), http.StatusText(nw.Status()))
		if sz := nw.Size(); sz > 0 {
			l.Printf(r.Context(), "  %v bytes written", sz)
		}
		if t := router.Time(r.Context()); !t.IsZero() {
			l.Printf(r.Context(), "  duration=%vms", time.Since(t).Milliseconds())
		}
	}
}
