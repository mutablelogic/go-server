package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	server "github.com/mutablelogic/go-server"
	logger "github.com/mutablelogic/go-server/pkg/logger"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Main is the main entry point for all commands. It parses command-line arguments and then dispatches to the appropriate command handler.
func Main[T any](cmds T, description, version string) error {
	var globals struct {
		global
		Cmds T `embed:""`
	}
	globals.Cmds = cmds

	// Get executable name
	if exe, err := os.Executable(); err != nil {
		return err
	} else {
		globals.execName = filepath.Base(exe)
		globals.version = version
		globals.description = description
	}

	// Parse command-line arguments
	kongctx := kong.Parse(&globals,
		kong.Name(globals.execName),
		kong.Description(description),
		kong.Vars{
			"version":         globals.version,
			"EXECUTABLE_NAME": globals.execName,
			"ENV_NAME":        strings.ToUpper(globals.execName),
		},
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Create logger
	var level slog.LevelVar
	if globals.Verbose {
		level.Set(logger.LevelTrace)
	} else if globals.Debug {
		level.Set(logger.LevelDebug)
	}
	if IsTerminal() {
		globals.logger = slog.New(logger.NewTermHandler(os.Stderr, &level))
	} else {
		globals.logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: &level}))
	}

	// Create the context and cancel function
	globals.ctx, globals.cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer globals.cancel()

	// Load defaults
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	} else if err := globals.init(filepath.Join(cacheDir, globals.execName, "defaults.json")); err != nil {
		return err
	}

	// Open Telemetry
	if globals.OTel.Endpoint != "" {
		provider, err := otel.NewProvider(globals.OTel.Endpoint, globals.OTel.Header, globals.OTel.Name)
		if err != nil {
			return err
		}
		defer otel.ShutdownProvider(context.Background())

		// Store tracer for creating spans
		globals.tracer = provider.Tracer(globals.OTel.Name)
	}

	// Bind the global context to the server.Cmd interface for command Run() methods.
	kongctx.BindTo(&globals.global, (*server.Cmd)(nil))

	// Call the Run() method of the selected parsed command.
	return kongctx.Run()
}

// IsTerminal reports whether os.Stderr is an interactive terminal.
func IsTerminal() bool {
	fi, err := os.Stderr.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}
