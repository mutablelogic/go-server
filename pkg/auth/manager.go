package auth

import (
	"context"
	"errors"

	// Packages
	pg "github.com/djthorpe/go-pg"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	conn pg.PoolConn
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new auth manager, with a root user
func NewManager(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*Manager, error) {
	self := new(Manager)
	self.conn = conn.With("schema", schema.SchemaName).(pg.PoolConn)

	_, err := apply(opt...)
	if err != nil {
		return nil, err
	}

	// TODO: Process options

	// If the schema does not exist, then bootstrap it
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		if exists, err := pg.SchemaExists(ctx, conn, schema.SchemaName); err != nil {
			return err
		} else if !exists {
			return schema.Bootstrap(ctx, conn)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}
