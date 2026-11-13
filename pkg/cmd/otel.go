package cmd

import (
	"context"
	"log/slog"

	// Packages
	gootel "go.opentelemetry.io/otel"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// setOTelErrorHandler reports OpenTelemetry's own errors, which have no
// caller, such as a metrics callback which fails, or an export, through
// the logger, at its level, and to its OpenTelemetry log export, rather
// than the log package's standard error
func setOTelErrorHandler(log *slog.Logger) {
	gootel.SetErrorHandler(gootel.ErrorHandlerFunc(func(err error) {
		log.ErrorContext(context.Background(), "opentelemetry", "error", err)
	}))
}
