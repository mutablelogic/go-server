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

type GroupCommands struct {
	Groups      GroupListCommand   `cmd:"" group:"LDAP" help:"List groups"`
	Group       GroupGetCommand    `cmd:"" group:"LDAP" help:"Get group"`
	CreateGroup GroupCreateCommand `cmd:"" group:"LDAP" help:"Create group"`
	DeleteGroup GroupDeleteCommand `cmd:"" group:"LDAP" help:"Delete group"`
}

type GroupListCommand struct {
	schema.ObjectListRequest
}

type GroupGetCommand struct {
	Group string `arg:"" help:"Group name"`
}

type GroupCreateCommand struct {
	GroupGetCommand
	Attr []string `arg:"" help:"attribute=value,value,..."`
}

type GroupDeleteCommand struct {
	GroupGetCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd GroupListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		groups, err := provider.ListGroups(ctx, client.WithFilter(cmd.Filter), client.WithAttr(cmd.Attr...), client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print groups
		fmt.Println(groups)
		return nil
	})
}

func (cmd GroupGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		group, err := provider.GetGroup(ctx, cmd.Group)
		if err != nil {
			return err
		}

		// Print group
		fmt.Println(group)
		return nil
	})
}

func (cmd GroupDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteGroup(ctx, cmd.Group)
	})
}

func (cmd GroupCreateCommand) Run(ctx server.Cmd) error {
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

		// Create group
		group, err := provider.CreateGroup(ctx, schema.Object{
			DN:     cmd.Group,
			Values: attrs,
		})
		if err != nil {
			return err
		}

		// Print group
		fmt.Println(group)
		return nil
	})
}
