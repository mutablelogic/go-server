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

		// Print the method and request
		l.Printf(ctx, "%v %v", r.Method, r.URL)

		// Call the next handler
		next(nw, r)

		// Print the response
		l.Printf(ctx, "  %v: %q", nw.Status(), http.StatusText(nw.Status()))
		if sz := nw.Size(); sz > 0 {
			l.Printf(ctx, "  %v bytes written", sz)
		}
		if t := router.Time(r.Context()); !t.IsZero() {
			l.Printf(ctx, "  duration=%vms", time.Since(t).Milliseconds())
		}
	}
}
