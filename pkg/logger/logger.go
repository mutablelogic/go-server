package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	// Packages
	server "github.com/mutablelogic/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Logger struct {
	*slog.Logger
	level slog.LevelVar
}

var _ server.Logger = (*Logger)(nil)

type Format uint

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Text Format = iota
	JSON
	Term
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(w io.Writer, f Format, debug bool) *Logger {
	l := &Logger{}
	if debug {
		l.level.Set(slog.LevelDebug)
	}
	var handler slog.Handler
	switch f {
	case JSON:
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: &l.level,
		})
	case Text:
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: &l.level,
		})
	case Term:
		handler = &TermHandler{
			Writer: w,
			level:  &l.level,
		}
	}
	l.Logger = slog.New(handler)
	return l
}

// SetDebug dynamically switches the log level between debug and info.
func (t *Logger) SetDebug(v bool) {
	if v {
		t.level.Set(slog.LevelDebug)
	} else {
		t.level.Set(slog.LevelInfo)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Emit a debugging message
func (t *Logger) Debug(ctx context.Context, args ...any) {
	t.log(ctx, slog.LevelDebug, fmt.Sprint(args...))
}

// Emit a debugging message with formatting
func (t *Logger) Debugf(ctx context.Context, f string, args ...any) {
	t.log(ctx, slog.LevelDebug, fmt.Sprintf(f, args...))
}

// Emit an informational message
func (t *Logger) Print(ctx context.Context, args ...any) {
	t.log(ctx, level(args...), fmt.Sprint(args...))
}

// Emit an informational message with formatting
func (t *Logger) Printf(ctx context.Context, f string, args ...any) {
	t.log(ctx, level(args...), fmt.Sprintf(f, args...))
}

// Append structured data to the log in key-value pairs
// where the key is a string and the value is any type
func (t *Logger) With(kv ...any) server.Logger {
	return &Logger{
		Logger: t.Logger.With(kv...),
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Emit a debugging message
func (t *Logger) log(ctx context.Context, level slog.Level, v string) {
	t.Logger.Log(ctx, level, v)
}

// Return error level if any arguments are errors
func level(args ...any) slog.Level {
	for _, arg := range args {
		if _, ok := arg.(error); ok {
			return slog.LevelError
		}
	}
	return slog.LevelInfo
}
