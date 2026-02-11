package main

import (
	"encoding/json"
	"fmt"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ResourcesCommands struct {
	ListResources           ListResourcesCommand           `cmd:"" name:"resources" help:"List all resources." group:"RESOURCE"`
	GetResourceInstance     GetResourceInstanceCommand     `cmd:"" name:"get" help:"Get a resource instance and its state." group:"RESOURCE"`
	CreateResourceInstance  CreateResourceInstanceCommand  `cmd:"" name:"create" help:"Create a new resource instance." group:"RESOURCE"`
	UpdateResourceInstance  UpdateResourceInstanceCommand  `cmd:"" name:"update" help:"Update a resource instance (plan or apply)." group:"RESOURCE"`
	DestroyResourceInstance DestroyResourceInstanceCommand `cmd:"" name:"destroy" help:"Destroy a resource instance." group:"RESOURCE"`
	OpenAPI                 OpenAPICommand                 `cmd:"" name:"openapi" help:"Retrieve the OpenAPI spec from a running server." group:"RESOURCE"`
}

type DestroyResourceInstanceCommand struct {
	Name    string `arg:"" help:"Instance name (e.g. \"httpserver.main\")"`
	Cascade bool   `flag:"" help:"Also destroy all instances that depend on this one." default:"false"`
}

type ListResourcesCommand struct {
	schema.ListResourcesRequest
}

type CreateResourceInstanceCommand struct {
	Names []string `arg:"" help:"Instance names as resource.label (e.g. \"httpserver.main httprouter.files\")"`
}

type GetResourceInstanceCommand struct {
	Name string `arg:"" help:"Instance name (e.g. \"httpserver.main\")"`
}

type UpdateResourceInstanceCommand struct {
	Name  string            `arg:"" help:"Instance name (e.g. \"httpserver.main\")"`
	Set   map[string]string `flag:"" help:"Set attribute values (e.g. --set port=8080)." optional:""`
	Del   []string          `flag:"" help:"Delete attributes (e.g. --del title)." optional:""`
	Apply bool              `flag:"" help:"Apply the planned changes." default:"false"`
}

type OpenAPICommand struct {
	Router string `arg:"" optional:"" default:"httprouter.main" help:"Router instance name (default: httprouter.main)"`
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

func (cmd *GetResourceInstanceCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	result, err := client.GetResourceInstance(ctx.ctx, cmd.Name)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func (cmd *CreateResourceInstanceCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// Create each instance; on error, destroy the ones already created
	var created []schema.CreateResourceInstanceResponse
	for _, name := range cmd.Names {
		result, err := client.CreateResourceInstance(ctx.ctx, schema.CreateResourceInstanceRequest{Name: name})
		if err != nil {
			// Roll back: destroy in reverse order
			for i := len(created) - 1; i >= 0; i-- {
				_, _ = client.DestroyResourceInstance(ctx.ctx, created[i].Instance.Name, false)
			}
			return err
		}
		created = append(created, *result)
	}

	// Print results
	for _, r := range created {
		fmt.Println(r)
	}
	return nil
}

func (cmd *DestroyResourceInstanceCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	result, err := client.DestroyResourceInstance(ctx.ctx, cmd.Name, cmd.Cascade)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func (cmd *UpdateResourceInstanceCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// Build attributes from --set and --del flags
	var attrs schema.State
	if len(cmd.Set) > 0 || len(cmd.Del) > 0 {
		attrs = make(schema.State, len(cmd.Set)+len(cmd.Del))
		for k, v := range cmd.Set {
			attrs[k] = parseValue(v)
		}
		for _, k := range cmd.Del {
			attrs[k] = nil
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

func (cmd *OpenAPICommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	raw, err := client.GetOpenAPI(ctx.ctx, cmd.Router)
	if err != nil {
		return err
	}

	// Pretty-print the JSON
	indented, err := json.MarshalIndent(json.RawMessage(raw), "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(indented))
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// parseValue attempts to interpret a string as JSON (arrays, objects,
// numbers, booleans, null). If it isn't valid JSON, the original
// string is returned as-is.
func parseValue(s string) any {
	if len(s) == 0 {
		return s
	}
	switch s[0] {
	case '[', '{':
		var v any
		if json.Unmarshal([]byte(s), &v) == nil {
			return v
		}
	}
	return s
}
