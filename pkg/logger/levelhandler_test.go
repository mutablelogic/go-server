package logger

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

type stubHandler struct {
	enabled bool
	handled []slog.Record
}

func (h *stubHandler) Enabled(context.Context, slog.Level) bool {
	return h.enabled
}

func (h *stubHandler) Handle(_ context.Context, record slog.Record) error {
	h.handled = append(h.handled, record.Clone())
	return nil
}

func (h *stubHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *stubHandler) WithGroup(string) slog.Handler {
	return h
}

func TestLevelHandlerEnabled(t *testing.T) {
	var level slog.LevelVar
	level.Set(LevelWarn)

	handler := NewLevelHandler(&stubHandler{enabled: true}, &level)
	if handler.Enabled(t.Context(), LevelInfo) {
		t.Fatal("expected info level to be disabled")
	}
	if !handler.Enabled(t.Context(), LevelError) {
		t.Fatal("expected error level to be enabled")
	}
}

func TestLevelHandlerHandleSkipsBelowThreshold(t *testing.T) {
	var level slog.LevelVar
	level.Set(LevelWarn)

	stub := &stubHandler{enabled: true}
	handler := NewLevelHandler(stub, &level)

	if err := handler.Handle(t.Context(), slog.NewRecord(time.Now(), LevelInfo, "ignored", 0)); err != nil {
		t.Fatalf("unexpected handle error: %v", err)
	}

	if len(stub.handled) != 0 {
		t.Fatalf("expected no records to be handled, got %d", len(stub.handled))
	}

	if err := handler.Handle(t.Context(), slog.NewRecord(time.Now(), LevelError, "sent", 0)); err != nil {
		t.Fatalf("unexpected handle error: %v", err)
	}

	if len(stub.handled) != 1 {
		t.Fatalf("expected 1 record to be handled, got %d", len(stub.handled))
	}
	if got := stub.handled[0].Message; got != "sent" {
		t.Fatalf("unexpected message %q", got)
	}
}
