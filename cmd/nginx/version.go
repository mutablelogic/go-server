package main

import (
	"fmt"

	nginx "github.com/mutablelogic/go-server/pkg/nginx/client"
)

// Version returns nginx version, build information and uptime duration
func Version(flags *Flags, client *nginx.Client) error {
	version, err := client.Health()
	if err != nil {
		return err
	}
	fmt.Println(version)
	return nil
}
