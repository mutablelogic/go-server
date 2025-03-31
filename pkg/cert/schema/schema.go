package schema

import (
	"context"

	// Packages
	pg "github.com/djthorpe/go-pg"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SchemaName = "cert"
)

const (
	// Maximum number of names to return in a list query
	NameListLimit = 100
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Bootstrap creates the schema and tables for the certificate manager
// and returns an error if it fails. It is expected that this function
// will be called within a transaction
func Bootstrap(ctx context.Context, conn pg.Conn) error {
	// Create the schema
	if err := pg.SchemaCreate(ctx, conn, SchemaName); err != nil {
		return err
	}
	// Create the tables
	if err := bootstrapName(ctx, conn); err != nil {
		return err
	}
	if err := bootstrapCert(ctx, conn); err != nil {
		return err
	}

	// Commit the transaction
	return nil
}
