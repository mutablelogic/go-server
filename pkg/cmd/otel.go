package cmd

import (
	"context"
	"log/slog"
	"sync"
	"time"

	// Packages
	gootel "go.opentelemetry.io/otel"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// otelErrors reports OpenTelemetry's own errors through a logger, each
// message once, and then, while it's repeated, once each otelErrorInterval,
// with how many times it was, such as an exporter's, which fails on each
// export while its collector can't be reached
type otelErrors struct {
	log *slog.Logger
	now func() time.Time

	mu     sync.Mutex
	errors map[string]*otelError
}

// otelError is when a message was last logged, and how many times it was
// repeated since
type otelError struct {
	logged   time.Time
	repeated int
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// otelErrorInterval is how often a message which is repeated is logged
	otelErrorInterval = 5 * time.Minute

	// otelErrorMessages is the most messages which are remembered, beyond
	// which the one logged least recently is forgotten
	otelErrorMessages = 64
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// setOTelErrorHandler reports OpenTelemetry's own errors, which have no
// caller, such as a metrics callback which fails, or an export, through
// the logger, at its level, rather than the log package's standard error,
// each message once, and then once each otelErrorInterval while it's
// repeated, with how many times it was. The logger is the console's, not
// one which exports to OpenTelemetry, as an export which fails would report
// its error by exporting it.
func setOTelErrorHandler(log *slog.Logger) {
	handler := &otelErrors{log: log, now: time.Now, errors: make(map[string]*otelError)}
	gootel.SetErrorHandler(gootel.ErrorHandlerFunc(handler.handle))
}

// handle logs the error, unless its message was logged within the
// interval, when it's counted
func (h *otelErrors) handle(err error) {
	if err == nil {
		return
	}
	message, now := err.Error(), h.now()
	h.mu.Lock()
	e, exists := h.errors[message]
	if exists && now.Sub(e.logged) < otelErrorInterval {
		e.repeated++
		h.mu.Unlock()
		return
	}
	repeated := 0
	if exists {
		repeated = e.repeated
	} else {
		h.forget()
	}
	h.errors[message] = &otelError{logged: now}
	h.mu.Unlock()

	if repeated > 0 {
		h.log.ErrorContext(context.Background(), "opentelemetry", "error", err, "repeated", repeated)
	} else {
		h.log.ErrorContext(context.Background(), "opentelemetry", "error", err)
	}
}

// forget removes the message logged least recently, when as many as are
// remembered are, which must be locked
func (h *otelErrors) forget() {
	if len(h.errors) < otelErrorMessages {
		return
	}
	var oldest string
	for message, e := range h.errors {
		if oldest == "" || e.logged.Before(h.errors[oldest].logged) {
			oldest = message
		}
	}
	delete(h.errors, oldest)
}
