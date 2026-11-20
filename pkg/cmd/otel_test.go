package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

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

// A message which is repeated, such as an exporter's, while its collector
// can't be reached, is logged once each interval, with how many times it was
func Test_OTelErrorHandlerRepeated(t *testing.T) {
	var buf bytes.Buffer
	now := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	h := &otelErrors{log: slog.New(slog.NewJSONHandler(&buf, nil)), now: func() time.Time { return now }, errors: make(map[string]*otelError)}
	down := errors.New(`traces export: Post "https://api.honeycomb.io/v1/traces": dial tcp: lookup api.honeycomb.io: no such host`)
	for range 10 {
		h.handle(down)
		now = now.Add(time.Second)
	}
	h.handle(errors.New("another"))
	assert.Equal(t, 1, strings.Count(buf.String(), "no such host"), "once, while it's repeated")
	assert.Equal(t, 1, strings.Count(buf.String(), "another"), "another message is logged")

	// Once the interval has passed, it's logged again, with how many times
	// it was repeated
	now = now.Add(otelErrorInterval)
	h.handle(down)
	assert.Equal(t, 2, strings.Count(buf.String(), "no such host"))
	assert.Contains(t, buf.String(), `"repeated":9`)

	// As many messages are remembered as there can be, forgetting the oldest
	for i := range otelErrorMessages + 10 {
		h.handle(fmt.Errorf("error %d", i))
	}
	assert.LessOrEqual(t, len(h.errors), otelErrorMessages)
}
