package cmd

import (
	"context"
	"fmt"

	// Packages
	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/cert/client"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type NameCommands struct {
	Names      NameListCommand   `cmd:"" group:"CERTIFICATES" help:"List names"`
	Name       NameGetCommand    `cmd:"get" group:"CERTIFICATES" help:"Get a name"`
	CreateName NameCreateCommand `cmd:"create" group:"CERTIFICATES" help:"Create a new name"`
	DeleteName NameDeleteCommand `cmd:"delete" group:"CERTIFICATES" help:"Delete a name"`
	UpdateName NameUpdateCommand `cmd:"update" group:"CERTIFICATES" help:"Update a name"`
}

type NameListCommand struct {
	schema.NameListRequest
}

type NameMeta struct {
	Org           *string `name:"org" description:"Organization name"`
	Unit          *string `name:"unit" description:"Organization unit"`
	Country       *string `name:"country" description:"Country name"`
	City          *string `name:"city" description:"City name"`
	State         *string `name:"state" description:"State name"`
	StreetAddress *string `name:"address" description:"Street address"`
	PostalCode    *string `name:"zip" description:"Postal code"`
}

type NameCreateCommand struct {
	NameMeta
}

type NameGetCommand struct {
	Id uint64 `arg:"" name:"id" description:"Name ID"`
}

type NameDeleteCommand struct {
	NameGetCommand
}

type NameUpdateCommand struct {
	NameGetCommand
	NameMeta
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd NameGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		name, err := provider.GetName(ctx, cmd.Id)
		if err != nil {
			return err
		}

		// Print name
		fmt.Println(name)
		return nil
	})
}

func (cmd NameDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteName(ctx, cmd.Id)
	})
}

func (cmd NameListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		names, err := provider.ListNames(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print names
		fmt.Println(names)
		return nil
	})
}

func (cmd NameCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		name, err := provider.CreateName(ctx, schema.NameMeta{
			Org:           cmd.Org,
			Unit:          cmd.Unit,
			Country:       cmd.Country,
			City:          cmd.City,
			State:         cmd.State,
			StreetAddress: cmd.StreetAddress,
			PostalCode:    cmd.PostalCode,
		})
		if err != nil {
			return err
		}

		// Print name
		fmt.Println(name)
		return nil
	})
}

func (cmd NameUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		name, err := provider.UpdateName(ctx, cmd.Id, schema.NameMeta{
			Org:           cmd.Org,
			Unit:          cmd.Unit,
			Country:       cmd.Country,
			City:          cmd.City,
			State:         cmd.State,
			StreetAddress: cmd.StreetAddress,
			PostalCode:    cmd.PostalCode,
		})
		if err != nil {
			return err
		}

		// Print name
		fmt.Println(name)
		return nil
	})
}
