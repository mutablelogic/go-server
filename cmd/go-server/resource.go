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
	UpdateResourceInstance UpdateResourceInstanceCommand `cmd:"" name:"update" help:"Update a resource instance (plan or apply)." group:"RESOURCE"`
}

type ListResourcesCommand struct {
	schema.ListResourcesRequest
}

type CreateResourceInstanceCommand struct {
	schema.CreateResourceInstanceRequest
}

type UpdateResourceInstanceCommand struct {
	Name  string            `arg:"" help:"Instance name (e.g. \"httpserver-01\")"` // instance name
	Set   map[string]string `flag:"" help:"Set attribute values (e.g. --set port=8080)." optional:""`
	Apply bool              `flag:"" help:"Apply the planned changes." default:"false"`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *ListResourcesCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// Get resources
	result, err := client.ListResources(ctx.ctx, cmd.ListResourcesRequest)
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

func (cmd *UpdateResourceInstanceCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// Build attributes from --set flags
	var attrs schema.State
	if len(cmd.Set) > 0 {
		attrs = make(schema.State, len(cmd.Set))
		for k, v := range cmd.Set {
			attrs[k] = v
		}
	}

	// Update instance (plan only unless --apply is set)
	result, err := client.UpdateResourceInstance(ctx.ctx, cmd.Name, schema.UpdateResourceInstanceRequest{
		Attributes: attrs,
		Apply:      cmd.Apply,
	})
	if err != nil {
		return err
	}

	// Print result
	fmt.Println(result)
	return nil
}
