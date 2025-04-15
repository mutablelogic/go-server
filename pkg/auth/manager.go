package auth

import (
	"context"
	"errors"

	// Packages
	pg "github.com/djthorpe/go-pg"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	conn pg.PoolConn
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new auth manager, with a root user
func New(ctx context.Context, conn pg.PoolConn, opt ...Opt) (*Manager, error) {
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

// Create a new user
func (manager *Manager) CreateUser(ctx context.Context, meta schema.UserMeta) (*schema.User, error) {
	var user schema.User
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := conn.Get(ctx, &user, schema.UserName(types.PtrString(meta.Name))); errors.Is(err, pg.ErrNotFound) {
			// OK
		} else if err != nil {
			return err
		} else {
			return httpresponse.ErrConflict.With("user already exists")
		}
		if err := conn.Insert(ctx, &user, meta); err != nil {
			return err
		}
		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}
	// Return success
	return &user, nil
}

// Create a new user, or update if the name already exists
func (manager *Manager) ReplaceUser(ctx context.Context, meta schema.UserMeta) (*schema.User, error) {
	var user schema.User
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := conn.Insert(ctx, &user, meta); err != nil {
			return err
		}
		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}
	// Return success
	return &user, nil
}

// Get a user
func (manager *Manager) GetUser(ctx context.Context, name string) (*schema.User, error) {
	var user schema.User
	if err := manager.conn.Get(ctx, &user, schema.UserName(name)); err != nil {
		return nil, httperr(err)
	}
	// Return success
	return &user, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}
