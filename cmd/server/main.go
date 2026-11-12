// Command server runs a HTTP server with go-server's APIs, such as the
// broadcaster's stream of changes, and is a client of them
package main

import (
	"fmt"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	broadcaster "github.com/mutablelogic/go-server/broadcaster"
	broadcastercmd "github.com/mutablelogic/go-server/broadcaster/cmd"
	config "github.com/mutablelogic/go-server/config"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	version "github.com/mutablelogic/go-server/pkg/version"
	errgroup "golang.org/x/sync/errgroup"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	ServerCommands                // server run
	broadcastercmd.ClientCommands // server changes
	servercmd.OpenAPICommands     // server openapi
}

type ServerCommands struct {
	Run RunServer `cmd:"" name:"run" help:"Run the server." group:"SERVER"`
}

type RunServer struct {
	servercmd.RunServer
	Broadcaster broadcastercmd.Flags `embed:"" prefix:"broadcaster."`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const description = "server runs a HTTP server with go-server's APIs, such as a stream of changes, and is a client of them"

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := servercmd.Main(CLI{}, description, version.Version()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

// Run runs the server, with its configuration, at /config.json, and the
// broadcaster's stream of changes, until it's interrupted, when the
// broadcaster stops, so its streams end
func (runner *RunServer) Run(ctx server.Cmd) error {
	runner.Register(func(router *httprouter.Router) error {
		return config.RegisterHandler(router, config.New(ctx, router, runner.OpenAPI))
	})
	return runner.Broadcaster.WithBroadcaster(ctx, &runner.RunServer, func(b *broadcaster.Broadcaster) error {
		group, groupCtx := errgroup.WithContext(ctx.Context())
		group.Go(func() error {
			return b.Run(groupCtx)
		})
		group.Go(func() error {
			return runner.RunServer.Run(ctx.WithContext(groupCtx))
		})
		return group.Wait()
	})
}
