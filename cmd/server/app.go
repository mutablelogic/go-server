package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	client "github.com/mutablelogic/go-client"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	Endpoint string `env:"ENDPOINT" default:"http://localhost/" help:"Service endpoint"`
	Debug    bool   `help:"Enable debug output"`
	Trace    bool   `help:"Enable trace output"`

	vars   kong.Vars `kong:"-"` // Variables for kong
	ctx    context.Context
	cancel context.CancelFunc
}

var _ server.Cmd = (*Globals)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewApp(app Globals, vars kong.Vars) (*Globals, error) {
	// Set the vars
	app.vars = vars

	// Create the context
	// This context is cancelled when the process receives a SIGINT or SIGTERM
	app.ctx, app.cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Return the app
	return &app, nil
}

func (app *Globals) Close() error {
	app.cancel()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (app *Globals) Context() context.Context {
	return app.ctx
}

func (app *Globals) GetDebug() bool {
	return app.Debug || app.Trace
}

func (app *Globals) GetEndpoint(paths ...string) *url.URL {
	url, err := url.Parse(app.Endpoint)
	if err != nil {
		return nil
	}
	for _, path := range paths {
		url.Path = types.JoinPath(url.Path, os.Expand(path, func(key string) string {
			return app.vars[key]
		}))
	}
	return url
}

func (app *Globals) GetClientOpts() []client.ClientOpt {
	opts := []client.ClientOpt{}

	// Trace mode
	if app.Debug || app.Trace {
		opts = append(opts, client.OptTrace(os.Stderr, app.Trace))
	}

	// Append user agent
	source := version.GitSource
	version := version.GitTag
	if source == "" {
		source = "go-server"
	}
	if version == "" {
		version = "v0.0.0"
	}
	opts = append(opts, client.OptUserAgent(fmt.Sprintf("%v/%v", filepath.Base(source), version)))

	// Return options
	return opts
}
