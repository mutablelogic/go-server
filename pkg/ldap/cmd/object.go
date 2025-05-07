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

type ObjectCommands struct {
	Objects      ObjectListCommand   `cmd:"" group:"LDAP" help:"List objects"`
	Object       ObjectGetCommand    `cmd:"" group:"LDAP" help:"Get object"`
	CreateObject ObjectCreateCommand `cmd:"" group:"LDAP" help:"Create object"`
	UpdateObject ObjectUpdateCommand `cmd:"" group:"LDAP" help:"Update object attributes"`
	DeleteObject ObjectDeleteCommand `cmd:"" group:"LDAP" help:"Delete object"`
}

type ObjectListCommand struct {
	schema.ObjectListRequest
}

type ObjectCreateCommand struct {
	ObjectGetCommand
	Attr []string `arg:"" help:"attribute=value,value,..."`
}

type ObjectGetCommand struct {
	DN string `arg:"" help:"Distingushed Name"`
}

type ObjectDeleteCommand struct {
	ObjectGetCommand
}

type ObjectUpdateCommand struct {
	ObjectCreateCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd ObjectListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		objects, err := provider.ListObjects(ctx, client.WithFilter(cmd.Filter), client.WithAttr(cmd.Attr...), client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print objects
		fmt.Println(objects)
		return nil
	})
}

func (cmd ObjectCreateCommand) Run(ctx server.Cmd) error {
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

		// Create object
		object, err := provider.CreateObject(ctx, schema.Object{
			DN:     cmd.DN,
			Values: attrs,
		})
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(object)
		return nil
	})
}

func (cmd ObjectUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Decode attributes
		attrs := url.Values{}
		for _, attr := range cmd.Attr {
			parts := strings.SplitN(attr, "=", 2)
			name := parts[0]
			if len(parts) == 1 {
				attrs[name] = []string{}
			} else {
				attrs[name] = strings.Split(parts[1], ",")
			}
		}

		// Update object
		object, err := provider.UpdateObject(ctx, schema.Object{
			DN:     cmd.DN,
			Values: attrs,
		})
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(object)
		return nil
	})
}

func (cmd ObjectGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		object, err := provider.GetObject(ctx, cmd.DN)
		if err != nil {
			return err
		}

		// Print object
		fmt.Println(object)
		return nil
	})
}

func (cmd ObjectDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteObject(ctx, cmd.DN)
	})
}
