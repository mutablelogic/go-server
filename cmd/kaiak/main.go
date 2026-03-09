package main

import (
	"fmt"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	httpclient "github.com/mutablelogic/go-server/pkg/provider/httpclient"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	ServerCommands
	ResourcesCommands
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	cli := new(CLI)
	if err := cmd.Main(cli, "kaiak command line interface", version.Version()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// CLIENT

func clientFor(ctx server.Cmd) (*httpclient.Client, error) {
	endpoint, opts, err := ctx.ClientEndpoint()
	if err != nil {
		return nil, err
	}
	return httpclient.New(endpoint, opts...)
}
