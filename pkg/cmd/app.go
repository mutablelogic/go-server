package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"os/user"
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

type app struct {
	Endpoint string `env:"ENDPOINT" default:"http://localhost/" help:"Service endpoint"`
	Debug    bool   `help:"Enable debug output"`
	Trace    bool   `help:"Enable trace output"`

	kong   *kong.Context `kong:"-"` // Kong parser
	vars   kong.Vars     `kong:"-"` // Variables for kong
	ctx    context.Context
	cancel context.CancelFunc
}

var _ server.Cmd = (*app)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(commands any, description string) (server.Cmd, error) {
	app := new(app)

	app.kong = kong.Parse(commands,
		kong.Name(execName()),
		kong.Description(description),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Embed(app),
		kong.Vars{
			"HOST": hostName(),
			"USER": userName(),
		},
	)

	// Create the context
	// This context is cancelled when the process receives a SIGINT or SIGTERM
	app.ctx, app.cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Return the app
	return app, nil
}

func (app *app) Run() error {
	app.kong.BindTo(app, (*server.Cmd)(nil))
	err := app.kong.Run()
	app.cancel()
	return err
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (app *app) Context() context.Context {
	return app.ctx
}

func (app *app) GetDebug() server.DebugLevel {
	if app.Debug {
		return server.Debug
	}
	if app.Trace {
		return server.Trace
	}
	return server.None
}

func (app *app) GetEndpoint(paths ...string) *url.URL {
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

func (app *app) GetClientOpts() []client.ClientOpt {
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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func hostName() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}

func userName() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.Username
}

func execName() string {
	// The name of the executable
	name, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Base(name)
}
