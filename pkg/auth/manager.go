package auth

import (
	"context"
	"errors"
	"strings"

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

	// Create and/or update the root user
	if _, err := replaceUser(ctx, self.conn, schema.UserMeta{
		Name: types.StringPtr(schema.RootUserName),
		Desc: types.StringPtr("Root user"),
		Scope: []string{
			schema.RootUserScope,
		},
		Meta: map[string]any{},
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
	if err := isRootUser(types.PtrString(meta.Name), "replace"); err != nil {
		return nil, err
	}
	return replaceUser(ctx, manager.conn, meta)
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

// Archive or delete a user
func (manager *Manager) DeleteUser(ctx context.Context, name string, force bool) (*schema.User, error) {
	var user schema.User
	if err := isRootUser(name, "delete"); err != nil {
		return nil, err
	}
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get the user to check current status
		if err := conn.Get(ctx, &user, schema.UserName(name)); err != nil {
			return httperr(err)
		}

		// Archive or delete the user
		switch schema.UserStatus(user.Status) {
		case schema.UserStatusArchived:
			if force {
				if err := manager.conn.Delete(ctx, &user, schema.UserName(name)); err != nil {
					return httperr(err)
				} else {
					return nil
				}
			}
		case schema.UserStatusLive:
			if force {
				if err := manager.conn.Delete(ctx, &user, schema.UserName(name)); err != nil {
					return httperr(err)
				} else {
					return nil
				}
			} else {
				if err := manager.conn.Update(ctx, &user, schema.UserName(name), schema.UserStatusArchived); err != nil {
					return httperr(err)
				} else {
					return nil
				}
			}
		}

		// If we get here, there was a conflict
		return httpresponse.ErrConflict.Withf("user cannot be archived or deleted")
	}); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &user, nil
}

// Update a user
func (manager *Manager) UpdateUser(ctx context.Context, name string, meta schema.UserMeta) (*schema.User, error) {
	var user schema.User
	if err := isRootUser(name, "update"); err != nil {
		return nil, err
	}
	if err := manager.conn.Update(ctx, &user, schema.UserName(name), meta); err != nil {
		return nil, httperr(err)
	}
	// Return success
	return &user, nil
}

// List users
func (manager *Manager) ListUsers(ctx context.Context, req schema.UserListRequest) (*schema.UserListResponse, error) {
	var response schema.UserListResponse
	if err := manager.conn.List(ctx, &response, req); err != nil {
		return nil, httperr(err)
	} else {
		response.UserListRequest = req
	}
	// Return success
	return &response, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func isRootUser(name, op string) error {
	if name := strings.ToLower(strings.TrimSpace(name)); name == schema.RootUserName {
		return httpresponse.ErrConflict.Withf("cannot %s root user", op)
	}
	return nil
}

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}

func replaceUser(ctx context.Context, conn pg.Conn, meta schema.UserMeta) (*schema.User, error) {
	var user schema.User
	if err := conn.Insert(ctx, &user, meta); err != nil {
		return nil, httperr(err)
	}
	return &user, nil
}
