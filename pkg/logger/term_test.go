package logger

import (
	"bytes"
	"context"
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
