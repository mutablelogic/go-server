package auth_test

import (
	"context"
	"testing"

	// Packages
	test "github.com/djthorpe/go-pg/pkg/test"
	auth "github.com/mutablelogic/go-server/pkg/auth"
	schema "github.com/mutablelogic/go-server/pkg/auth/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
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

func Test_Auth_001(t *testing.T) {
	assert := assert.New(t)
	conn := conn.Begin(t)
	defer conn.Close()

	// Create a new database manager
	manager, err := auth.New(context.TODO(), conn)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// Create a new user
	t.Run("CreateUser", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test"),
			Desc:  types.StringPtr("test user"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)
	})

	// Replace a new user (that doesn't exist)
	t.Run("ReplaceUser1", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test2"),
			Desc:  types.StringPtr("test user"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.ReplaceUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)
	})

	// Replace a user (that does exist)
	t.Run("ReplaceUser2", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test2"),
			Desc:  types.StringPtr("test user"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.ReplaceUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		meta2 := schema.UserMeta{
			Name:  meta.Name,
			Desc:  types.StringPtr("test user 2"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user2, err := manager.ReplaceUser(context.TODO(), meta2)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta2, user2.UserMeta)
	})

	// Replace root user
	t.Run("ReplaceUser3", func(t *testing.T) {
		_, err := manager.ReplaceUser(context.TODO(), schema.UserMeta{
			Name: types.StringPtr("root"),
		})
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})

	// Get a non-existent user
	t.Run("GetUser1", func(t *testing.T) {
		_, err := manager.GetUser(context.TODO(), "non_existent")
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Get a user
	t.Run("GetUser2", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test3"),
			Desc:  types.StringPtr("test user"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		user2, err := manager.GetUser(context.TODO(), types.PtrString(meta.Name))
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user2.UserMeta)
	})

	// Delete a user
	t.Run("DeleteUser1", func(t *testing.T) {
		_, err := manager.DeleteUser(context.TODO(), "non_existent", true)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Delete a user
	t.Run("DeleteUser2", func(t *testing.T) {
		_, err := manager.DeleteUser(context.TODO(), "root", true)
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})

	// Delete a user
	t.Run("DeleteUser3", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test4"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		user2, err := manager.DeleteUser(context.TODO(), types.PtrString(meta.Name), true)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user2.UserMeta)

		_, err = manager.GetUser(context.TODO(), types.PtrString(meta.Name))
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Archive a user
	t.Run("ArchiveUser1", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test4"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		user2, err := manager.DeleteUser(context.TODO(), types.PtrString(meta.Name), false)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("archived", user2.Status)
		assert.Equal(meta, user2.UserMeta)
	})

	// Unarchive a user
	t.Run("ArchiveUser2", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test8"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		user2, err := manager.DeleteUser(context.TODO(), types.PtrString(meta.Name), false)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("archived", user2.Status)
		assert.Equal(meta, user2.UserMeta)

		user3, err := manager.UnarchiveUser(context.TODO(), types.PtrString(meta.Name))
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("live", user3.Status)
		assert.Equal(meta, user3.UserMeta)
	})

	// Update a user
	t.Run("UpdateUser1", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test5"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		meta2 := schema.UserMeta{
			Name:  types.StringPtr("test6"),
			Desc:  types.StringPtr("test user"),
			Scope: []string{"a", "b"},
			Meta: map[string]any{
				"c": float64(1.0),
				"d": float64(2.0),
			},
		}
		user2, err := manager.UpdateUser(context.TODO(), types.PtrString(user.Name), meta2)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta2, user2.UserMeta)
	})

	// Update a user
	t.Run("UpdateUser2", func(t *testing.T) {
		_, err := manager.UpdateUser(context.TODO(), "root", schema.UserMeta{})
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})

	// Update a user
	t.Run("UpdateUser3", func(t *testing.T) {
		meta := schema.UserMeta{
			Name:  types.StringPtr("test7"),
			Scope: []string{},
			Meta:  map[string]any{},
		}
		user, err := manager.CreateUser(context.TODO(), meta)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(meta, user.UserMeta)

		_, err = manager.UpdateUser(context.TODO(), types.PtrString(user.Name), schema.UserMeta{})
		assert.ErrorIs(err, httpresponse.ErrBadRequest)
	})

	// List users
	t.Run("ListUsers1", func(t *testing.T) {
		response, err := manager.ListUsers(context.TODO(), schema.UserListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		t.Log(response)
	})

	// List users
	t.Run("ListUsers2", func(t *testing.T) {
		response, err := manager.ListUsers(context.TODO(), schema.UserListRequest{
			Scope:  types.StringPtr("root"),
			Status: types.StringPtr("live"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		t.Log(response)
	})

}
