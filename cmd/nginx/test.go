package main

import (
	nginx "github.com/mutablelogic/go-server/pkg/nginx/client"
)

// Test will test the nginx configuration
func Test(flags *Flags, client *nginx.Client) error {
	return client.Test()
}

// Reopen will reopen nginx log files
func Reopen(flags *Flags, client *nginx.Client) error {
	return client.Reopen()
}

// Reload will reload nginx configuration files
func Reload(flags *Flags, client *nginx.Client) error {
	return client.Reload()
}
