package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mutablelogic/go-server/pkg/client"
	"github.com/mutablelogic/go-server/pkg/httpserver/router"
)

func main() {
	flags, err := NewFlags("cli", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Add command-line options
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

	// Create the client
	client, err := client.New(opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Router client
	// Add commands for rouuter:
	//   router - returns a list of services, description
	//   router <service> - returns a list of methods for a service
	router := router.NewClient(client, flags.Router())
	if router == nil {
		fmt.Fprintln(os.Stderr, "Unable to create router client")
		os.Exit(1)
	}

	// Obtain routes
	routes, err := router.Services()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if data, err := json.MarshalIndent(routes, "", "  "); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		fmt.Println(string(data))
	}
}
