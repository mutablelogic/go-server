package pgmanager_test

import (
	"context"
	"testing"
	"time"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgmanager "github.com/mutablelogic/go-server/pkg/pgmanager"
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

// Global connection variable
var conn test.Conn

// Start up a container and test the pool
func TestMain(m *testing.M) {
	test.Main(m, &conn)
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Test_Manager_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new database manager
	manager, err := pgmanager.New(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// List roles
	t.Run("ListRoles", func(t *testing.T) {
		roles, err := manager.ListRoles(context.TODO(), schema.RoleListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(roles)
		assert.Equal(len(roles.Body), int(roles.Count))
	})

	// Get roles
	t.Run("GetRoles", func(t *testing.T) {
		roles, err := manager.ListRoles(context.TODO(), schema.RoleListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		for _, role := range roles.Body {
			role2, err := manager.GetRole(context.TODO(), role.Name)
			if !assert.NoError(err) {
				t.FailNow()
			}
			assert.Equal(role, *role2)
		}
	})

	// Not Found Role
	t.Run("GetNonExistentRole", func(t *testing.T) {
		_, err := manager.GetRole(context.TODO(), "non_existing_role")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Create Role
	t.Run("CreateRole", func(t *testing.T) {
		meta := schema.RoleMeta{
			Name:                   "role",
			Superuser:              types.BoolPtr(true),
			Inherit:                types.BoolPtr(true),
			CreateRoles:            types.BoolPtr(true),
			CreateDatabases:        types.BoolPtr(true),
			Replication:            types.BoolPtr(true),
			ConnectionLimit:        types.Uint64Ptr(10),
			BypassRowLevelSecurity: types.BoolPtr(true),
			Login:                  types.BoolPtr(true),
			Password:               types.StringPtr(t.Name()),
			Expires:                types.TimePtr(time.Now()),
		}
		role, err := manager.CreateRole(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Some fields will change from write->read
		meta.Password = role.Password
		meta.Expires = role.Expires

		// Check equality
		assert.Equal(role.RoleMeta, meta)
	})

	// Create Role With Memberships
	t.Run("CreateRoleMember", func(t *testing.T) {
		meta1 := schema.RoleMeta{
			Name: "role1",
		}
		meta2 := schema.RoleMeta{
			Name:   "role2",
			Groups: []string{"role1"},
		}
		role1, err := manager.CreateRole(context.TODO(), meta1)
		if !assert.NoError(err) {
			t.FailNow()
		}

		role2, err := manager.CreateRole(context.TODO(), meta2)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Check equality
		assert.Equal(role1.RoleMeta.Groups, meta1.Groups)
		assert.Equal(role2.RoleMeta.Groups, meta2.Groups)
	})

	// Update a role
	t.Run("UpdateRole", func(t *testing.T) {
		meta := schema.RoleMeta{
			Name: "role3",
		}
		role, err := manager.CreateRole(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		role2, err := manager.UpdateRole(context.TODO(), role.Name, schema.RoleMeta{
			Name:      "role4",
			Superuser: types.BoolPtr(true),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(role.Name, "role3")
		assert.Equal(role2.Name, "role4")
		assert.Equal(types.PtrBool(role2.Superuser), true)
	})
}

func Test_Manager_002(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new database manager
	manager, err := pgmanager.New(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// List databases
	t.Run("ListDatabases", func(t *testing.T) {
		databases, err := manager.ListDatabases(context.TODO(), schema.DatabaseListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(databases)
		assert.Equal(len(databases.Body), int(databases.Count))
	})

	// Get databases
	t.Run("GetDatabases", func(t *testing.T) {
		databases, err := manager.ListDatabases(context.TODO(), schema.DatabaseListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		for _, database := range databases.Body {
			database2, err := manager.GetDatabase(context.TODO(), database.Name)
			if !assert.NoError(err) {
				t.FailNow()
			}
			assert.Equal(database, *database2)
		}
	})

	// Not Found Database
	t.Run("GetNonExistentDatabase", func(t *testing.T) {
		_, err := manager.GetDatabase(context.TODO(), "non_existing_database")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Create Database
	t.Run("CreateDatabase", func(t *testing.T) {
		meta := schema.DatabaseMeta{
			Name: "database",
		}
		database, err := manager.CreateDatabase(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Some fields will change from write->read
		meta.Owner = database.Owner
		meta.Acl = database.Acl

		// Check equality
		assert.Equal(database.DatabaseMeta, meta)
	})

	// Delete Database
	t.Run("DeleteDatabase", func(t *testing.T) {
		database, err := manager.CreateDatabase(context.TODO(), schema.DatabaseMeta{
			Name: "database2",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		database2, err := manager.DeleteDatabase(context.TODO(), database.Name, false)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(database, database2)

		_, err = manager.GetDatabase(context.TODO(), database.Name)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Delete Non-Existing Database
	t.Run("DeleteNonExistentDatabase", func(t *testing.T) {
		_, err := manager.DeleteDatabase(context.TODO(), "non_existing_database", false)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Rename a database
	t.Run("UpdateDatabaseRename", func(t *testing.T) {
		database, err := manager.CreateDatabase(context.TODO(), schema.DatabaseMeta{
			Name: "database3",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		database2, err := manager.UpdateDatabase(context.TODO(), database.Name, schema.DatabaseMeta{
			Name: "database4",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(database2.Name, "database4")

		_, err = manager.GetDatabase(context.TODO(), database.Name)
		assert.ErrorIs(err, httpresponse.ErrNotFound)

		_, err = manager.GetDatabase(context.TODO(), database2.Name)
		if !assert.NoError(err) {
			t.FailNow()
		}
	})

	// Change database owner
	t.Run("UpdateDatabaseOwner", func(t *testing.T) {
		role1, err := manager.CreateRole(context.TODO(), schema.RoleMeta{
			Name: "role5",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		role2, err := manager.CreateRole(context.TODO(), schema.RoleMeta{
			Name: "role6",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		database, err := manager.CreateDatabase(context.TODO(), schema.DatabaseMeta{
			Name:  "database5",
			Owner: role1.Name,
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(database.Owner, role1.Name)

		database2, err := manager.UpdateDatabase(context.TODO(), database.Name, schema.DatabaseMeta{
			Owner: role2.Name,
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(database2.Owner, role2.Name)
	})

	// Add ACL
	t.Run("AddDatabasePublicAcl", func(t *testing.T) {
		database, err := manager.CreateDatabase(context.TODO(), schema.DatabaseMeta{
			Name: "database6",
			Acl: schema.ACLList{&schema.ACLItem{
				Role: "PUBLIC",
				Priv: []string{"ALL"},
			}},
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		roles := database.Acl.Find("PUBLIC")
		if assert.NotNil(roles) {
			assert.Equal([]string{"CREATE", "TEMPORARY", "CONNECT"}, roles.Priv)
		}
	})

	// Update ACL
	t.Run("UpdateDatabasePublicAcl", func(t *testing.T) {
		database, err := manager.CreateDatabase(context.TODO(), schema.DatabaseMeta{
			Name: "database7",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		database2, err := manager.UpdateDatabase(context.TODO(), database.Name, schema.DatabaseMeta{
			Acl: schema.ACLList{&schema.ACLItem{
				Role: "PUBLIC",
				Priv: []string{"ALL"},
			}},
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		roles := database2.Acl.Find("PUBLIC")
		if assert.NotNil(roles) {
			assert.Equal([]string{"CREATE", "TEMPORARY", "CONNECT"}, roles.Priv)
		}
	})
}

func Test_Manager_003(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new database manager
	manager, err := pgmanager.New(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// List schemas
	t.Run("ListSchemas", func(t *testing.T) {
		schemas, err := manager.ListSchemas(context.TODO(), schema.SchemaListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(schemas)
		assert.Equal(len(schemas.Body), int(schemas.Count))
	})

	// Get schemas
	t.Run("GetSchemas", func(t *testing.T) {
		schemas, err := manager.ListSchemas(context.TODO(), schema.SchemaListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		for _, schema := range schemas.Body {
			schema2, err := manager.GetSchema(context.TODO(), schema.Database, schema.Name)
			if !assert.NoError(err) {
				t.FailNow()
			}
			assert.Equal(schema, *schema2)
			t.Log(schema2)
		}
	})

	// Not Found Schema
	t.Run("GetNonExistentSchema", func(t *testing.T) {
		_, err := manager.GetSchema(context.TODO(), "postgres", "non_existing_schema")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Create Schema
	t.Run("CreateSchema", func(t *testing.T) {
		role, err := manager.CreateRole(context.TODO(), schema.RoleMeta{
			Name: "role7",
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		meta := schema.SchemaMeta{
			Name:  "schema",
			Owner: role.Name,
			Acl:   schema.ACLList{},
		}
		schema, err := manager.CreateSchema(context.TODO(), "postgres", meta)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Check equality
		assert.Equal(schema.SchemaMeta, meta)
	})
}

func Test_Manager_004(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new database manager
	manager, err := pgmanager.New(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// List objects
	t.Run("ListObjects", func(t *testing.T) {
		objects, err := manager.ListObjects(context.TODO(), schema.ObjectListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(objects)
		assert.Equal(len(objects.Body), int(objects.Count))
		t.Log(objects)
	})
}

func Test_Manager_005(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new database manager
	manager, err := pgmanager.New(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// List tablespaces
	t.Run("ListTablespaces", func(t *testing.T) {
		objects, err := manager.ListTablespaces(context.TODO(), schema.TablespaceListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(objects)
		assert.Equal(len(objects.Body), int(objects.Count))
		t.Log(objects)
	})

	// Get non-existent tablespace
	t.Run("GetTablespaces1", func(t *testing.T) {
		_, err := manager.GetTablespace(context.TODO(), "non_existing_tablespace")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Get tablespaces
	t.Run("GetTablespaces1", func(t *testing.T) {
		objects, err := manager.ListTablespaces(context.TODO(), schema.TablespaceListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		for _, tablespace := range objects.Body {
			tablespace2, err := manager.GetTablespace(context.TODO(), types.PtrString(tablespace.Name))
			if !assert.NoError(err) {
				t.FailNow()
			}
			assert.Equal(tablespace, *tablespace2)
		}
	})
}
