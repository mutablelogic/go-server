package pgmanager

import (
	"context"
	"errors"
	"slices"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	conn pg.PoolConn
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new database manager
func New(ctx context.Context, conn pg.PoolConn) (*Manager, error) {
	self := new(Manager)
	self.conn = conn.With("schema", schema.CatalogSchema).(pg.PoolConn)

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (manager *Manager) ListRoles(ctx context.Context, req schema.RoleListRequest) (*schema.RoleList, error) {
	var list schema.RoleList
	if err := manager.conn.List(ctx, &list, req); err != nil {
		return nil, httperr(err)
	} else {
		return &list, nil
	}
}

func (manager *Manager) ListDatabases(ctx context.Context, req schema.DatabaseListRequest) (*schema.DatabaseList, error) {
	var list schema.DatabaseList
	if err := manager.conn.List(ctx, &list, req); err != nil {
		return nil, httperr(err)
	} else {
		return &list, nil
	}
}

func (manager *Manager) ListSchemas(ctx context.Context, req schema.SchemaListRequest) (*schema.SchemaList, error) {
	var list schema.SchemaList
	if err := manager.conn.List(ctx, &list, req); err != nil {
		return nil, httperr(err)
	} else {
		return &list, nil
	}
}

func (manager *Manager) GetRole(ctx context.Context, name string) (*schema.Role, error) {
	var role schema.Role
	if err := manager.conn.Get(ctx, &role, schema.RoleName(name)); err != nil {
		return nil, httperr(err)
	}
	return &role, nil
}

func (manager *Manager) GetDatabase(ctx context.Context, name string) (*schema.Database, error) {
	var database schema.Database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(name)); err != nil {
		return nil, httperr(err)
	}
	return &database, nil
}

func (manager *Manager) CreateRole(ctx context.Context, meta schema.RoleMeta) (*schema.Role, error) {
	var role schema.Role
	if err := manager.conn.Insert(ctx, nil, meta); err != nil {
		return nil, httperr(err)
	} else if err := manager.conn.Get(ctx, &role, schema.RoleName(meta.Name)); err != nil {
		return nil, httperr(err)
	}
	return &role, nil
}

func (manager *Manager) CreateDatabase(ctx context.Context, meta schema.Database) (*schema.Database, error) {
	var database schema.Database

	// Create the database - cannot be done in a transaction
	if err := manager.conn.Insert(ctx, nil, meta); err != nil {
		return nil, httperr(err)
	}

	// TODO: Set ACL's - this must be done in a transaction
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Return success
		return nil
	}); err != nil {
		return nil, errors.Join(httperr(err), manager.conn.Delete(ctx, nil, schema.DatabaseName(meta.Name)))
	}

	// Get the database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(meta.Name)); err != nil {
		return nil, httperr(err)
	}
	return &database, nil
}

func (manager *Manager) DeleteRole(ctx context.Context, name string) (*schema.Role, error) {
	var role schema.Role
	if err := manager.conn.Get(ctx, &role, schema.RoleName(name)); err != nil {
		return nil, httperr(err)
	} else if err := manager.conn.Delete(ctx, nil, schema.RoleName(name)); err != nil {
		return nil, httperr(err)
	}
	return &role, nil
}

func (manager *Manager) DeleteDatabase(ctx context.Context, name string) (*schema.Database, error) {
	var database schema.Database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(name)); err != nil {
		return nil, httperr(err)
	} else if err := manager.conn.Delete(ctx, nil, schema.DatabaseName(name)); err != nil {
		return nil, httperr(err)
	}
	return &database, nil
}

func (manager *Manager) UpdateRole(ctx context.Context, name string, meta schema.RoleMeta) (*schema.Role, error) {
	var role schema.Role

	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get the role and memberships
		if err := manager.conn.Get(ctx, &role, schema.RoleName(name)); err != nil {
			return err
		}

		// Update the name if it's different
		if meta.Name != "" && name != meta.Name {
			if err := conn.Update(ctx, nil, schema.RoleName(meta.Name), schema.RoleName(name)); err != nil {
				return err
			}
		} else {
			meta.Name = name
		}

		// Update the rest of the metadata
		if err := conn.Update(ctx, nil, meta, meta); err != nil {
			return err
		}

		// Update the group memberships
		if meta.Groups != nil {
			// Remove the old roles
			for _, oldrole := range role.Groups {
				if !slices.Contains(meta.Groups, oldrole) {
					if err := schema.RevokeGroupMembership(ctx, conn, oldrole, meta.Name); err != nil {
						return err
					}
				}
			}
			// Add the new roles
			for _, newrole := range meta.Groups {
				if !slices.Contains(role.Groups, newrole) {
					if err := schema.GrantGroupMembership(ctx, conn, newrole, meta.Name); err != nil {
						return err
					}
				}
			}
		}

		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}

	// Get the role
	if err := manager.conn.Get(ctx, &role, schema.RoleName(meta.Name)); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &role, nil
}

func (manager *Manager) UpdateDatabase(ctx context.Context, name string, meta schema.Database) (*schema.Database, error) {
	var database schema.Database

	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get the database and ACL's
		if err := manager.conn.Get(ctx, &database, schema.DatabaseName(name)); err != nil {
			return err
		}

		// Update the name if it's different
		if meta.Name != "" && name != meta.Name {
			if err := conn.Update(ctx, nil, schema.DatabaseName(meta.Name), schema.DatabaseName(name)); err != nil {
				return err
			}
		} else {
			meta.Name = name
		}

		// Update the rest of the metadata
		if err := conn.Update(ctx, nil, meta, meta); err != nil {
			return err
		}

		// TODO Update ACL's

		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}

	// Get the database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(meta.Name)); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &database, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}
