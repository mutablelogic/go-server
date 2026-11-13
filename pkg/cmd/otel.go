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
// the logger, at its level, rather than the log package's standard error.
// The logger is the console's, not one which exports to OpenTelemetry, as
// an export which fails would report its error by exporting it.
func setOTelErrorHandler(log *slog.Logger) {
	gootel.SetErrorHandler(gootel.ErrorHandlerFunc(func(err error) {
		log.ErrorContext(context.Background(), "opentelemetry", "error", err)
	}))
}
