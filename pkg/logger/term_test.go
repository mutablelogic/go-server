package logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestTermHandlerAttrs(t *testing.T) {
	var buf bytes.Buffer
	level := new(slog.LevelVar)
	logger := slog.New(NewTermHandler(&buf, level).WithAttrs([]slog.Attr{slog.String("worker", "a<b>")}))
	logger.InfoContext(context.Background(), "message", "url", "https://example.com/a.mp3?x=1&y=2")

	// The attributes are JSON, on the same line as the message, whose
	// characters which are special in HTML aren't escaped
	line := buf.String()
	if !strings.HasSuffix(line, `{"url":"https://example.com/a.mp3?x=1&y=2","worker":"a<b>"}`+"\n") {
		t.Errorf("unexpected line %q", line)
	}
	if strings.Count(line, "\n") != 1 {
		t.Errorf("expected one line, got %q", line)
	}
}

func TestTermHandlerNoAttrs(t *testing.T) {
	var buf bytes.Buffer
	slog.New(NewTermHandler(&buf, new(slog.LevelVar))).Info("message")
	if strings.Contains(buf.String(), "{") {
		t.Errorf("expected no attributes, got %q", buf.String())
	}
}

// opaque is an error with no exported fields, which encodes as {}
type opaque struct{ code int }

func (e opaque) Error() string { return fmt.Sprint("failed with ", e.code) }

func TestTermHandlerErrors(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewTermHandler(&buf, new(slog.LevelVar)))
	logger.Error("opentelemetry", "error", opaque{code: 7}, "wrapped", fmt.Errorf("export: %w", errors.New("timeout")), slog.Group("export", "error", opaque{code: 8}))

	// An error is its message, in a group too, rather than {}
	line := buf.String()
	if !strings.HasSuffix(line, `{"error":"failed with 7","export":{"error":"failed with 8"},"wrapped":"export: timeout"}`+"\n") {
		t.Errorf("unexpected line %q", line)
	}
}
