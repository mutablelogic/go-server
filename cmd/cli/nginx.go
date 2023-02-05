package main

import (
	"encoding/json"
	"errors"
	"fmt"

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
			{MinArgs: 1, MaxArgs: 1, Description: "Return nginx status", Fn: NginxStatus(nginx, flags)},
			{MinArgs: 2, MaxArgs: 2, Name: "test", Description: "Test nginx configuration", Fn: NginxTest(nginx, flags)},
			{MinArgs: 2, MaxArgs: 2, Name: "reopen", Description: "Reopen nginx log files", Fn: NginxReopen(nginx, flags)},
			{MinArgs: 2, MaxArgs: 2, Name: "reload", Description: "Reload nginx configuration", Fn: NginxReload(nginx, flags)},
		},
	})

	// Return success
	return cmd, nil
}

func NginxStatus(nginx *nginx.Client, flags *Flags) CommandFn {
	return func() error {
		// Obtain status
		status, err := nginx.Health()
		if err != nil {
			return err
		}
		if data, err := json.MarshalIndent(status, "", "  "); err != nil {
			return err
		} else {
			fmt.Println(string(data))
		}
		// Return success
		return nil
	}
}

func NginxReopen(nginx *nginx.Client, flags *Flags) CommandFn {
	return func() error {
		return nginx.Reopen()
	}
}

func NginxReload(nginx *nginx.Client, flags *Flags) CommandFn {
	return func() error {
		return nginx.Reload()
	}
}

func NginxTest(nginx *nginx.Client, flags *Flags) CommandFn {
	return func() error {
		return nginx.Test()
	}
}
