package logger

import (
	"context"
	"log/slog"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type LevelHandler struct {
	handler slog.Handler
	level   slog.Leveler
}

var _ slog.Handler = (*LevelHandler)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewLevelHandler(handler slog.Handler, level slog.Leveler) slog.Handler {
	if handler == nil || level == nil {
		return handler
	}
	return &LevelHandler{handler: handler, level: level}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (h *LevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level() && h.handler.Enabled(ctx, level)
}

func (h *LevelHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}
	return h.handler.Handle(ctx, record)
}

func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelHandler{handler: h.handler.WithAttrs(attrs), level: h.level}
}

func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return &LevelHandler{handler: h.handler.WithGroup(name), level: h.level}
}
