package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	// Packages
	client "github.com/mutablelogic/go-client"
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

func clientFor(ctx *cmd.Global) (*httpclient.Client, error) {
	endpoint, opts, err := clientEndpointFor(ctx)
	if err != nil {
		return nil, err
	}
	return httpclient.New(endpoint, opts...)
}

func clientEndpointFor(ctx *cmd.Global) (string, []client.ClientOpt, error) {
	scheme := "http"
	host, port, err := net.SplitHostPort(ctx.HTTP.Addr)
	if err != nil {
		return "", nil, err
	}
	if host == "" {
		host = "localhost"
	}
	portn, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", nil, err
	}
	if portn == 443 {
		scheme = "https"
	}

	opts := []client.ClientOpt{}
	if ctx.Debug || ctx.Verbose {
		opts = append(opts, client.OptTrace(os.Stderr, ctx.Verbose))
	}
	if ctx.Tracer() != nil {
		opts = append(opts, client.OptTracer(ctx.Tracer()))
	}
	if ctx.HTTP.Timeout > 0 {
		opts = append(opts, client.OptTimeout(ctx.HTTP.Timeout))
	}

	return fmt.Sprintf("%s://%s:%v%s", scheme, host, portn, ctx.HTTP.Prefix), opts, nil
}
