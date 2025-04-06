package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type TermHandler struct {
	slog.Handler
}

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
		level = colorize(cyan, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	}

	fmt.Println(
		colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		colorize(white, r.Message),
	)

	return nil
}

/////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}
