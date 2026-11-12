package cmd

import (
	"context"
	"encoding/json"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpclient "github.com/mutablelogic/go-server/broadcaster/httpclient"
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ChangeCommands are the commands which watch the broadcaster's changes
type ChangeCommands struct {
	Changes ChangesCmd `cmd:"" name:"changes" help:"Watch the changes, such as to the rows of the server's tables, as lines of JSON." group:"CHANGES"`
}

type ChangesCmd struct {
	schema.ChangeFilter `embed:""`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run prints each change as a line of JSON, until the command is interrupted
func (cmd *ChangesCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "broadcaster.ChangesCmd", func(ctx context.Context, client *httpclient.Client) error {
		encoder := json.NewEncoder(os.Stdout)
		return client.WatchChanges(ctx, cmd.ChangeFilter, func(change schema.Change) error {
			return encoder.Encode(change)
		})
	})
}
