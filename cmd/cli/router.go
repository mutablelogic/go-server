package main

import (
	"encoding/json"
	"errors"
	"fmt"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
	router "github.com/mutablelogic/go-server/pkg/httpserver/router"
)

func AppendRouter(cmd []Client, client *client.Client, flags *Flags) ([]Client, error) {
	// Create router client
	router := router.NewClient(client, flags.Router())
	if router == nil {
		return nil, errors.New("unable to create router client")
	}

	// Register commands for router
	cmd = append(cmd, Client{
		ns: "service",
		cmd: []Command{
			{MinArgs: 1, MaxArgs: 1, Description: "List services", Fn: ListServices(router, flags)},
			{MinArgs: 2, MaxArgs: 2, Description: "List routes for service", Fn: ListRoutes(router, flags)},
		},
	})

	// Return success
	return cmd, nil
}

func ListServices(router *router.Client, flags *Flags) CommandFn {
	return func() error {
		// Obtain services
		services, err := router.Services()
		if err != nil {
			return err
		}
		if data, err := json.MarshalIndent(services, "", "  "); err != nil {
			return err
		} else {
			fmt.Println(string(data))
		}
		// Return success
		return nil
	}
}

func ListRoutes(router *router.Client, flags *Flags) CommandFn {
	prefix := flags.Arg(1)
	return func() error {
		// Obtain routes
		services, err := router.Routes(prefix)
		if err != nil {
			return err
		}
		if data, err := json.MarshalIndent(services, "", "  "); err != nil {
			return err
		} else {
			fmt.Println(string(data))
		}
		// Return success
		return nil
	}
}
