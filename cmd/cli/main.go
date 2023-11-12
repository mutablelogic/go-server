package main

import (
	"flag"
	"fmt"
	"os"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
)

func main() {
	flags, err := NewFlags("cli", os.Args[1:])
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	// Add client options
	opts := []client.ClientOpt{}
	if flags.IsDebug() {
		opts = append(opts, client.OptTrace(os.Stderr, true))
	}
	if token := flags.Token(); token != "" {
		token := client.Token{Scheme: client.Bearer, Value: token}
		opts = append(opts, client.OptReqToken(token))
	}
	if url := flags.URL(); url != "" {
		opts = append(opts, client.OptEndpoint(url))
	}
	if timeout := flags.Timeout(); timeout > 0 {
		opts = append(opts, client.OptTimeout(timeout))
	}
	if flags.IsSkipVerify() {
		opts = append(opts, client.OptSkipVerify())
	}

	// Create the client
	client, err := client.New(opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Register commands
	var cmd []Client

	// Router
	cmd, err = AppendRouter(cmd, client, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Nginx
	cmd, err = AppendNginx(cmd, client, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Run command
	if err := Run(cmd, flags); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
