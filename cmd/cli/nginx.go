package main

import (
	"errors"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"
	nginx "github.com/mutablelogic/go-server/pkg/nginx"
)

func AppendNginx(cmd []Client, client *client.Client, flags *Flags) ([]Client, error) {
	// Create nginx client
	nginx := nginx.NewClient(client, flags.Nginx())
	if nginx == nil {
		return nil, errors.New("unable to create nginx client")
	}

	// Register commands for router
	cmd = append(cmd, Client{
		ns: "nginx",
		cmd: []Command{
			{MinArgs: 2, MaxArgs: 2, Name: "test", Description: "Test nginx configuration"},
			{MinArgs: 2, MaxArgs: 2, Name: "reopen", Description: "Reopen nginx log files"},
			{MinArgs: 2, MaxArgs: 2, Name: "reload", Description: "Reload nginx configuration"},
		},
	})

	// Return success
	return cmd, nil
}
