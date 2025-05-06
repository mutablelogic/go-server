package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/ldap/client"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type UserCommands struct {
	CreateUser UserCreateCommand `cmd:"" group:"LDAP" help:"Create user"`
}

type UserGetCommand struct {
	DN string `arg:"" help:"Username"`
}

type UserCreateCommand struct {
	UserGetCommand
	Attr []string `arg:"" help:"attribute=value,value,..."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd UserCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Decode attributes
		attrs := url.Values{}
		for _, attr := range cmd.Attr {
			parts := strings.SplitN(attr, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid attribute: %s", attr)
			}
			name := parts[0]
			if values := strings.Split(parts[1], ","); len(values) > 0 {
				attrs[name] = values
			}
		}

		// Create user
		user, err := provider.CreateUser(ctx, schema.Object{
			DN:     cmd.DN,
			Values: attrs,
		})
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(user)
		return nil
	})
}
