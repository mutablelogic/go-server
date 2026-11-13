package cmd

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
	gootel "go.opentelemetry.io/otel"
)

// OpenTelemetry's errors go to the logger
func Test_OTelErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	setOTelErrorHandler(slog.New(slog.NewJSONHandler(&buf, nil)))
	gootel.Handle(errors.New("RegisterConnectionMetrics: cannot scan NULL"))
	assert.Contains(t, buf.String(), `"level":"ERROR"`)
	assert.Contains(t, buf.String(), `"msg":"opentelemetry"`)
	assert.Contains(t, buf.String(), `"error":"RegisterConnectionMetrics: cannot scan NULL"`)
}
