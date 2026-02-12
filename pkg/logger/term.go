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
	level *slog.LevelVar
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
	var parts []any

	parts = append(parts, r.Time.Format(timeFormat))

	level := r.Level.String()
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
	parts = append(parts, level+":")

	//	label := ref.Label(ctx)
	//	if label != "" {
	//		parts = append(parts, label+":")
	//	}

	// Gather attributes
	var data []byte
	if data_, err := attrs(h.attrs, r); err != nil {
		return err
	} else {
		data = data_
	}

	parts = append(parts, colorize(white, r.Message), string(data))

	// Print the message, return any errors
	fmt.Fprintln(h.Writer, parts...)
	return nil
}

func (h *TermHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TermHandler{
		Writer: h.Writer,
		level:  h.level,
		attrs:  append(h.attrs, attrs...),
	}
}

func (h *TermHandler) WithGroup(name string) slog.Handler {
	// Groups not supported
	return &TermHandler{
		Writer: h.Writer,
		level:  h.level,
		attrs:  h.attrs,
	}
}

func (h *TermHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

/////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

func attrs(attrs []slog.Attr, r slog.Record) ([]byte, error) {
	if len(attrs) == 0 && r.NumAttrs() == 0 {
		return nil, nil
	}
	kv := make(map[string]any, len(attrs)+r.NumAttrs())
	for _, a := range attrs {
		if a.Equal(slog.Attr{}) {
			continue
		}
		kv[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		if a.Equal(slog.Attr{}) {
			return true
		}
		kv[a.Key] = a.Value.Any()
		return true
	})

	return json.Marshal(kv)
}
