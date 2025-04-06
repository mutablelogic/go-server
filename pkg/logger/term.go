package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type TermHandler struct {
	io.Writer
	slog.Level
	attrs []slog.Attr
}

var _ slog.Handler = (*TermHandler)(nil)

/////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	timeFormat = "[15:04:05.000]"
)

const (
	reset        = "\033[0m"
	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (h *TermHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = colorize(darkGray, level)
	case slog.LevelInfo:
		level = colorize(white, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	}

	var data []byte
	attrs := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		attrs[a.Key] = attrValue(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = attrValue(a)
		return true
	})
	if data_, err := json.MarshalIndent(attrs, "", "  "); err != nil {
		return err
	} else {
		data = data_
	}

	// Print the message, return any errors
	_, err := fmt.Fprintln(h.Writer,
		colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		colorize(white, r.Message),
		string(data),
	)
	return err
}

func (h *TermHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TermHandler{
		Writer: h.Writer,
		Level:  h.Level,
		attrs:  append(h.attrs, attrs...),
	}
}

func (h *TermHandler) WithGroup(name string) slog.Handler {
	// Groups not supported
	return &TermHandler{
		Writer: h.Writer,
		Level:  h.Level,
		attrs:  h.attrs,
	}
}

func (h *TermHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.Level
}

/////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

func attrValue(a slog.Attr) any {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return nil
	}
	return a.Value.Any()
}
