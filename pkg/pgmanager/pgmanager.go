package pgmanager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
	"github.com/mutablelogic/go-server/pkg/types"
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
	var offset, limit uint64

	// Set limit lower if request limit is lower
	limit = schema.SchemaListLimit
	if req.Limit != nil && types.PtrUint64(req.Limit) < limit {
		limit = types.PtrUint64(req.Limit)
	}

	// Allocate the body with capacity
	list.Body = make([]schema.Schema, 0, limit)

	// Iterate through all the databases
	if _, err := manager.withDatabases(ctx, func(database *schema.Database) error {
		// Filter by database
		if name := strings.TrimSpace(req.Database); name != "" && name != database.Name {
			return nil
		}

		// Iterate through all the schemas
		count, err := manager.withSchemas(ctx, database.Name, func(schema *schema.Schema) error {
			if offset >= req.Offset && uint64(len(list.Body)) < limit {
				list.Body = append(list.Body, *schema)
			}
			offset++
			return nil
		})
		if err != nil {
			return err
		}

		// Increment the count
		list.Count += count

		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}

	return &list, nil
}

func (manager *Manager) ListObjects(ctx context.Context, req schema.ObjectListRequest) (*schema.ObjectList, error) {
	var list schema.ObjectList
	if err := manager.conn.List(ctx, &list, req); err != nil {
		return nil, httperr(err)
	} else {
		return &list, nil
	}
}

func (manager *Manager) ListConnections(ctx context.Context, req schema.ConnectionListRequest) (*schema.ConnectionList, error) {
	var list schema.ConnectionList
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

func (manager *Manager) GetSchema(ctx context.Context, name string) (*schema.Schema, error) {
	var response schema.Schema
	database, name := schema.SchemaName(name).Split()
	if name == "" || database == "" {
		return nil, httpresponse.ErrBadRequest.With("database or schema is missing")
	}
	if err := manager.conn.Remote(database).With("as", schema.SchemaDef).Get(ctx, &response, schema.SchemaName(name)); err != nil {
		return nil, httperr(err)
	}
	return &response, nil
}

func (manager *Manager) GetConnection(ctx context.Context, pid uint64) (*schema.Connection, error) {
	var response schema.Connection
	if err := manager.conn.Get(ctx, &response, schema.ConnectionPid(pid)); err != nil {
		return nil, httperr(err)
	}
	return &response, nil
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

func (manager *Manager) CreateDatabase(ctx context.Context, meta schema.DatabaseMeta) (*schema.Database, error) {
	var database schema.Database

	// Create the database - cannot be done in a transaction
	if err := manager.conn.Insert(ctx, nil, meta); err != nil {
		return nil, httperr(err)
	}

	// Set ACL's - this can be done in a transaction
	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		for _, acl := range meta.Acl {
			if err := acl.GrantDatabase(ctx, conn, meta.Name); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		// Delete the database if there is an issue with ACL's
		return nil, errors.Join(httperr(err), manager.conn.Delete(ctx, nil, schema.DatabaseName(meta.Name)))
	}

	// Get the database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(meta.Name)); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &database, nil
}

func (manager *Manager) CreateSchema(ctx context.Context, meta schema.SchemaMeta) (*schema.Schema, error) {
	var response schema.Schema

	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Create the database - cannot be done in a transaction
		if err := manager.conn.Insert(ctx, nil, meta); err != nil {
			return err
		}

		// Set ACL's
		for _, acl := range meta.Acl {
			if err := acl.GrantSchema(ctx, conn, meta.Name); err != nil {
				return err
			}
		}

		// Get the schema
		if err := manager.conn.Get(ctx, &response, schema.SchemaName(meta.Name)); err != nil {
			return err
		}

		// Return success
		return nil
	}); err != nil {
		return nil, httperr(err)
	}

	return &response, nil
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

func (manager *Manager) DeleteDatabase(ctx context.Context, name string, force bool) (*schema.Database, error) {
	var database schema.Database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(name)); err != nil {
		return nil, httperr(err)
	} else if err := manager.conn.With("force", force).Delete(ctx, nil, schema.DatabaseName(name)); err != nil {
		return nil, httperr(err)
	}
	return &database, nil
}

func (manager *Manager) DeleteSchema(ctx context.Context, name string, force bool) (*schema.Schema, error) {
	var response schema.Schema

	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		if err := manager.conn.Get(ctx, &response, schema.SchemaName(name)); err != nil {
			return err
		}
		if err := manager.conn.With("force", force).Delete(ctx, nil, schema.SchemaName(name)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, httperr(err)
	}
	return &response, nil
}

func (manager *Manager) DeleteConnection(ctx context.Context, pid uint64) (*schema.Connection, error) {
	var connection schema.Connection
	if err := manager.conn.Delete(ctx, &connection, schema.ConnectionPid(pid)); err != nil {
		return nil, httperr(err)
	}
	return &connection, nil
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

func (manager *Manager) UpdateDatabase(ctx context.Context, name string, meta schema.DatabaseMeta) (*schema.Database, error) {
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

		// Update ACL's
		if meta.Acl != nil {
			for _, acl := range database.Acl {
				if role := meta.Acl.Find(acl.Role); role == nil {
					// Revoke the older privileges
					if err := acl.RevokeDatabase(ctx, conn, meta.Name); err != nil {
						return err
					}
				} else if slices.Equal(acl.Priv, role.Priv) {
					// No change
				} else if role.IsAll() {
					// Just grant
					if err := role.GrantDatabase(ctx, conn, meta.Name); err != nil {
						return err
					}
				} else {
					// Revoke
					for _, priv := range acl.Priv {
						if !slices.Contains(role.Priv, priv) {
							if err := acl.WithPriv(priv).RevokeDatabase(ctx, conn, meta.Name); err != nil {
								return err
							}
						}
					}
					// Grant
					for _, priv := range role.Priv {
						if !slices.Contains(acl.Priv, priv) {
							if err := acl.WithPriv(priv).GrantDatabase(ctx, conn, meta.Name); err != nil {
								return err
							}
						}
					}
				}
			}
			for _, acl := range meta.Acl {
				if role := database.Acl.Find(acl.Role); role == nil {
					// Create new privileges
					if err := acl.GrantDatabase(ctx, conn, meta.Name); err != nil {
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

	// Get the database
	if err := manager.conn.Get(ctx, &database, schema.DatabaseName(meta.Name)); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &database, nil
}

func (manager *Manager) UpdateSchema(ctx context.Context, name string, meta schema.SchemaMeta) (*schema.Schema, error) {
	var response schema.Schema

	if err := manager.conn.Tx(ctx, func(conn pg.Conn) error {
		// Get the schema and ACL's
		if err := manager.conn.Get(ctx, &response, schema.DatabaseName(name)); err != nil {
			return err
		}

		// Update the name if it's different
		if meta.Name != "" && name != meta.Name {
			if err := conn.Update(ctx, nil, schema.SchemaName(meta.Name), schema.SchemaName(name)); err != nil {
				return err
			}
		} else {
			meta.Name = name
		}

		// Update the rest of the metadata
		if err := conn.Update(ctx, nil, meta, meta); err != nil {
			return err
		}

		// Update ACL's
		if meta.Acl != nil {
			for _, acl := range response.Acl {
				if role := meta.Acl.Find(acl.Role); role == nil {
					// Revoke the older privileges
					if err := acl.RevokeDatabase(ctx, conn, meta.Name); err != nil {
						return err
					}
				} else if slices.Equal(acl.Priv, role.Priv) {
					// No change
				} else if role.IsAll() {
					// Just grant
					if err := role.GrantDatabase(ctx, conn, meta.Name); err != nil {
						return err
					}
				} else {
					fmt.Println("acl", acl.Priv, "=>", role.Priv)
					// Revoke
					for _, priv := range acl.Priv {
						if !slices.Contains(role.Priv, priv) {
							if err := acl.WithPriv(priv).RevokeDatabase(ctx, conn, meta.Name); err != nil {
								return err
							}
						}
					}
					// Grant
					for _, priv := range role.Priv {
						if !slices.Contains(acl.Priv, priv) {
							if err := acl.WithPriv(priv).GrantDatabase(ctx, conn, meta.Name); err != nil {
								return err
							}
						}
					}
				}
			}
			for _, acl := range meta.Acl {
				if role := response.Acl.Find(acl.Role); role == nil {
					// Create new privileges
					if err := acl.GrantDatabase(ctx, conn, meta.Name); err != nil {
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

	// Get the schema
	if err := manager.conn.Get(ctx, &response, schema.SchemaName(meta.Name)); err != nil {
		return nil, httperr(err)
	}

	// Return success
	return &response, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func httperr(err error) error {
	if errors.Is(err, pg.ErrNotFound) {
		return httpresponse.ErrNotFound.With(err)
	}
	return err
}

// Iterate through all the databases
func (manager *Manager) withDatabases(ctx context.Context, fn func(database *schema.Database) error) (uint64, error) {
	var req schema.DatabaseListRequest
	req.Offset = 0
	req.Limit = types.Uint64Ptr(schema.DatabaseListLimit)

	for {
		list, err := manager.ListDatabases(ctx, req)
		if err != nil {
			return 0, err
		}
		for _, database := range list.Body {
			if err := fn(&database); err != nil {
				return 0, err
			}
		}

		// Determine if the next page is over the count
		next := req.Offset + types.PtrUint64(req.Limit)
		if next >= list.Count {
			return list.Count, nil
		} else {
			req.Offset = next
		}
	}
}

// Iterate through all the schemas for a database
func (manager *Manager) withSchemas(ctx context.Context, database string, fn func(schema *schema.Schema) error) (uint64, error) {
	var req schema.SchemaListRequest
	req.Offset = 0
	req.Limit = types.Uint64Ptr(schema.SchemaListLimit)

	for {
		var list schema.SchemaList
		if err := manager.conn.Remote(database).With("as", schema.SchemaDef).List(ctx, &list, &req); err != nil {
			return 0, err
		}

		for _, schema := range list.Body {
			if err := fn(&schema); err != nil {
				return 0, err
			}
		}

		// Determine if the next page is over the count
		next := req.Offset + types.PtrUint64(req.Limit)
		if next >= list.Count {
			return list.Count, nil
		} else {
			req.Offset = next
		}
	}
}
