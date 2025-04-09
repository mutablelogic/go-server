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
		t.Log(roles)
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
		roles, err := manager.ListDatabases(context.TODO(), schema.DatabaseListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(roles)
		t.Log(roles)
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
		t.Log(schemas)
	})

}
