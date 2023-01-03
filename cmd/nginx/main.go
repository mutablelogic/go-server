/*
 * nginx - A client to restart, reload and test nginx
 */
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
	nginx "github.com/mutablelogic/go-server/pkg/nginx/client"
)

func main() {
	// Create flags
	flagset := NewFlags(filepath.Base(os.Args[0]))

	// Parse flags
	if err := flagset.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		} else {
			os.Exit(-1)
		}
	}

	// Check for version flag
	if flagset.GetVersion() {
		flagset.PrintVersion()
		os.Exit(0)
	}

	// Create the client, set options
	opts := []client.ClientOpt{}
	if token := flagset.GetToken(); token != "" {
		opts = append(opts, client.OptReqToken(token))
	}
	if flagset.GetDebug() {
		opts = append(opts, client.OptTrace(os.Stderr, true))
	}
	client, err := nginx.New(flagset.GetEndpoint(), opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Run client commands
	if fn, err := flagset.GetCommand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if err := fn(flagset, client); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
