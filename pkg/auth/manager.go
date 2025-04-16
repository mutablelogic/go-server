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
	self.conn = conn.With(
		"schema", schema.SchemaName,
		"algorithm", schema.AuthHashAlgorithm,
	).(pg.PoolConn)

	_, err := apply(opt...)
	if err != nil {
		return nil, err
	}

	// Bootstrap
	if err := self.conn.Tx(ctx, func(conn pg.Conn) error {
		if exists, err := pg.SchemaExists(ctx, conn, schema.SchemaName); err != nil {
			return err
		} else if !exists {
			if err := pg.SchemaCreate(ctx, conn, schema.SchemaName); err != nil {
				return err
			}
		}
		return schema.Bootstrap(ctx, conn)
	}); err != nil {
		return nil, err
	}

	// Create and/or update the root user
	// TODO: Create token and return it if that doesn't exist
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

// Create a new token for a user
func (manager *Manager) CreateToken(ctx context.Context, name string, meta schema.TokenMeta) (*schema.Token, error) {
	var token schema.Token
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		var user schema.User
		var value schema.TokenNew

		// Check for user and that user is live
		if err := conn.Get(ctx, &user, schema.UserName(name)); err != nil {
			return err
		} else if schema.UserStatus(user.Status) != schema.UserStatusLive {
			return httpresponse.ErrConflict.With("user is archived")
		}

		// Generate a new token
		if err := conn.Get(ctx, &value, value); err != nil {
			return err
		} else {
			value.User = types.PtrString(user.Name)
			value.TokenMeta = meta
		}

		// Insert the token, set the token value
		if err := conn.Insert(ctx, &token, value); err != nil {
			return err
		} else {
			token.Value = types.StringPtr(value.Value)
		}

		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}
	// Return success
	return &token, nil
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

// Get a token for a user
func (manager *Manager) GetToken(ctx context.Context, name string, id uint64) (*schema.Token, error) {
	var token schema.Token
	if err := manager.conn.Get(ctx, &token, schema.TokenId{User: name, Id: id}); err != nil {
		return nil, httperr(err)
	}
	// Return success
	return &token, nil
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
				return conn.Delete(ctx, &user, schema.UserName(name))
			}
		case schema.UserStatusLive:
			if force {
				return conn.Delete(ctx, &user, schema.UserName(name))
			} else {
				return conn.Update(ctx, &user, schema.UserName(name), schema.UserStatusArchived)
			}
		}

		// If we get here, there was a conflict
		return httpresponse.ErrConflict.With("user cannot be archived or deleted")
	}); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &user, nil
}

// Delete or archive token for a user
func (manager *Manager) DeleteToken(ctx context.Context, name string, id uint64, force bool) (*schema.Token, error) {
	var token schema.Token
	if err := isRootUser(name, "delete token for"); err != nil {
		return nil, err
	}
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Archive or delete the token
		if force {
			return conn.Delete(ctx, &token, schema.TokenId{User: name, Id: id})
		} else if err := conn.Update(ctx, &token, schema.TokenId{User: name, Id: id}, schema.TokenStatusArchived); err != nil {
			return err
		}

		// Re-read the token (status and desc fields may come from the user, not the token)
		return conn.Get(ctx, &token, schema.TokenId{User: name, Id: id})
	}); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &token, nil
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

// Update a token
func (manager *Manager) UpdateToken(ctx context.Context, name string, id uint64, meta schema.TokenMeta) (*schema.Token, error) {
	var token schema.Token
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := manager.conn.Update(ctx, &token, schema.TokenId{User: name, Id: id}, meta); err != nil {
			return err
		}

		// Re-read the token (status and desc fields may come from the user, not the token)
		return conn.Get(ctx, &token, schema.TokenId{User: token.User, Id: token.Id})
	}); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &token, nil
}

// Unarchive a user
func (manager *Manager) UnarchiveUser(ctx context.Context, name string) (*schema.User, error) {
	var user schema.User
	if err := isRootUser(name, "unarchive"); err != nil {
		return nil, err
	}
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get the user to check current status
		if err := conn.Get(ctx, &user, schema.UserName(name)); err != nil {
			return httperr(err)
		} else if schema.UserStatus(user.Status) != schema.UserStatusArchived {
			return httpresponse.ErrConflict.With("user is not archived")
		}

		// Unarchive the user
		return conn.Update(ctx, &user, schema.UserName(name), schema.UserStatusLive)
	}); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &user, nil
}

// List users
func (manager *Manager) ListUsers(ctx context.Context, req schema.UserListRequest) (*schema.UserList, error) {
	var response schema.UserList
	if err := manager.conn.List(ctx, &response, &req); err != nil {
		return nil, httperr(err)
	} else {
		response.UserListRequest = req
	}
	// Return success
	return &response, nil
}

// List tokens for a user
func (manager *Manager) ListTokens(ctx context.Context, name string, req schema.TokenListRequest) (*schema.TokenList, error) {
	var response schema.TokenList
	if err := manager.conn.With("user", name).List(ctx, &response, &req); err != nil {
		return nil, httperr(err)
	} else {
		response.TokenListRequest = req
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
