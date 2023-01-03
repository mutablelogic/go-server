package main

import (
	nginx "github.com/mutablelogic/go-server/pkg/nginx/client"
)

// Test will test the nginx configuration
func Test(flags *Flags, client *nginx.Client) error {
	return client.Test()
}
