package main

import (
	"encoding/json"
	"fmt"

	// Packages
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	httpclient "github.com/mutablelogic/go-server/pkg/provider/httpclient"
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

func (c *ListResourcesCommand) Run(ctx *cmd.Global) (err error) {
	client, err := newClient(ctx)
	if err != nil {
		return err
	}

	// Get resources
	result, err := client.ListResources(ctx.Context(), c.ListResourcesRequest)
	if err != nil {
		return err
	}

	// Print result
	fmt.Println(result)
	return nil
}

func (c *GetResourceInstanceCommand) Run(ctx *cmd.Global) (err error) {
	client, err := newClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.GetResourceInstance(ctx.Context(), c.Name)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func (c *CreateResourceInstanceCommand) Run(ctx *cmd.Global) (err error) {
	client, err := newClient(ctx)
	if err != nil {
		return err
	}

	// Create each instance; on error, destroy the ones already created
	var created []schema.CreateResourceInstanceResponse
	for _, name := range c.Names {
		result, err := client.CreateResourceInstance(ctx.Context(), schema.CreateResourceInstanceRequest{Name: name})
		if err != nil {
			// Roll back: destroy in reverse order
			for i := len(created) - 1; i >= 0; i-- {
				_, _ = client.DestroyResourceInstance(ctx.Context(), created[i].Instance.Name, false)
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

func (c *DestroyResourceInstanceCommand) Run(ctx *cmd.Global) (err error) {
	client, err := newClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.DestroyResourceInstance(ctx.Context(), c.Name, c.Cascade)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func (c *UpdateResourceInstanceCommand) Run(ctx *cmd.Global) (err error) {
	client, err := newClient(ctx)
	if err != nil {
		return err
	}

	// Build attributes from --set and --del flags
	var attrs schema.State
	if len(c.Set) > 0 || len(c.Del) > 0 {
		attrs = make(schema.State, len(c.Set)+len(c.Del))
		for k, v := range c.Set {
			attrs[k] = parseValue(v)
		}
		for _, k := range c.Del {
			attrs[k] = nil
		}
	}

	// Update instance (plan only unless --apply is set)
	result, err := client.UpdateResourceInstance(ctx.Context(), c.Name, schema.UpdateResourceInstanceRequest{
		Attributes: attrs,
		Apply:      c.Apply,
	})
	if err != nil {
		return err
	}

	// Print result
	fmt.Println(result)
	return nil
}

func (c *OpenAPICommand) Run(ctx *cmd.Global) (err error) {
	client, err := newClient(ctx)
	if err != nil {
		return err
	}

	raw, err := client.GetOpenAPI(ctx.Context(), c.Router)
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

// parseValue attempts to interpret a string as a JSON value (arrays,
// objects, numbers, booleans, null). If it isn't valid JSON or is a
// JSON string, the original string is returned as-is.
func parseValue(s string) any {
	if len(s) == 0 {
		return s
	}
	var v any
	if json.Unmarshal([]byte(s), &v) == nil {
		// Only use the parsed value when it is NOT a string —
		// bare strings should stay as Go strings, not go through
		// JSON round-tripping.
		if _, isStr := v.(string); !isStr {
			return v
		}
	}
	return s
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newClient(g *cmd.Global) (*httpclient.Client, error) {
	endpoint, opts, err := g.ClientEndpoint()
	if err != nil {
		return nil, err
	}
	return httpclient.New(endpoint, opts...)
}
