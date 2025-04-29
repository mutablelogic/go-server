package schema

import (
	"context"

	// Packages
	"github.com/djthorpe/go-pg"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SchemaName        = "auth"
	RootUserName      = "root"
	UserListLimit     = 50
	AuthHashAlgorithm = "sha256"
	APIPrefix         = "/auth/v1"
)

const (
	ScopeRoot      = "root"
	ScopeUserRead  = "mutablelogic/go-server/pkg/auth/user_read"
	ScopeUserWrite = "mutablelogic/go-server/pkg/auth/user_write"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Bootstrap creates the schema and tables for the auth manager
// and returns an error if it fails. It is expected that this function
// will be called within a transaction
func Bootstrap(ctx context.Context, conn pg.Conn) error {
	// Create the tables
	if err := bootstrapUser(ctx, conn); err != nil {
		return err
	}
	if err := bootstrapToken(ctx, conn); err != nil {
		return err
	}

	// Commit the transaction
	return nil
}
