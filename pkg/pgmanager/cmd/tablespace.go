package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	server "github.com/mutablelogic/go-server"
	client "github.com/mutablelogic/go-server/pkg/pgmanager/client"
	"github.com/mutablelogic/go-server/pkg/pgmanager/schema"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TablespaceCommands struct {
	Tablespaces      TablespaceListCommand   `cmd:"" group:"DATABASE" help:"List tablespaces"`
	Tablespace       TablespaceGetCommand    `cmd:"" group:"DATABASE" help:"Get tablespace"`
	CreateTablespace TablespaceCreateCommand `cmd:"" group:"DATABASE" help:"Create a new tablespace"`
	UpdateTablespace TablespaceUpdateCommand `cmd:"" group:"DATABASE" help:"Update a tablespace"`
	DeleteTablespace TablespaceDeleteCommand `cmd:"" group:"DATABASE" help:"Delete a tablespace"`
}

type TablespaceListCommand struct {
	schema.TablespaceListRequest
}

type TablespaceGetCommand struct {
	Name string `arg:"" help:"Tablespace name"`
}

type TablespaceCreateCommand struct {
	Location string `arg:"" help:"Tablespace path"`
	schema.TablespaceMeta
}

type TablespaceDeleteCommand struct {
	TablespaceGetCommand
}

type TablespaceUpdateCommand struct {
	Name string `arg:"" help:"Tablespace name"`
	schema.TablespaceMeta
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd TablespaceListCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		tablespaces, err := provider.ListTablespaces(ctx, client.WithOffsetLimit(cmd.Offset, cmd.Limit))
		if err != nil {
			return err
		}

		// Print tablespaces
		fmt.Println(tablespaces)
		return nil
	})
}

func (cmd TablespaceGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		tablespace, err := provider.GetTablespace(ctx, cmd.Name)
		if err != nil {
			return err
		}

		// Print tablespace
		fmt.Println(tablespace)
		return nil
	})
}

func (cmd TablespaceDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteTablespace(ctx, cmd.Name)
	})
}

func (cmd TablespaceCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Set name if not set
		if cmd.Name == nil {
			cmd.Name = types.StringPtr(filepath.Base(cmd.Location))
		}

		// Create tablespace
		tablespace, err := provider.CreateTablespace(ctx, cmd.TablespaceMeta, cmd.Location)
		if err != nil {
			return err
		}

		// Print database
		fmt.Println(tablespace)
		return nil
	})
}

func (cmd TablespaceUpdateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		// Perform request
		tablespace, err := provider.UpdateTablespace(ctx, cmd.Name, cmd.TablespaceMeta)
		if err != nil {
			return err
		}

		// Print tablespace
		fmt.Println(tablespace)
		return nil
	})
}
