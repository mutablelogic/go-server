package schema

import (
	"context"

	// Packages
	"github.com/djthorpe/go-pg"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SchemaName    = "auth"
	RootUserName  = "root"
	RootUserScope = "root"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Bootstrap creates the schema and tables for the auth manager
// and returns an error if it fails. It is expected that this function
// will be called within a transaction
func Bootstrap(ctx context.Context, conn pg.Conn) error {
	// Create the schema
	if err := pg.SchemaCreate(ctx, conn, SchemaName); err != nil {
		return err
	}
	// Create the tables
	if err := bootstrapUser(ctx, conn); err != nil {
		return err
	}

	// Commit the transaction
	return nil
}
