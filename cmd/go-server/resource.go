package main

import (
	"fmt"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ResourcesCommands struct {
	ListResources          ListResourcesCommand          `cmd:"" name:"resources" help:"List all resources." group:"RESOURCE"`
	CreateResourceInstance CreateResourceInstanceCommand `cmd:"" name:"create" help:"Create a new resource instance." group:"RESOURCE"`
}

type ListResourcesCommand struct {
}

type CreateResourceInstanceCommand struct {
	schema.CreateResourceInstanceRequest
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *ListResourcesCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// Get resources
	result, err := client.ListResources(ctx.ctx, schema.ListResourcesRequest{})
	if err != nil {
		return err
	}

	// Print result
	fmt.Println(result)
	return nil
}

func (cmd *CreateResourceInstanceCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// Get resources
	result, err := client.CreateResourceInstance(ctx.ctx, cmd.CreateResourceInstanceRequest)
	if err != nil {
		return err
	}

	// Print result
	fmt.Println(result)
	return nil
}
