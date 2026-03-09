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
// GLOBALS

const description = "Kaiak provides an infrastructure as code interface for composing and managing software resources."

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := cmd.Main(CLI{}, description, version.Version()); err != nil {
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
