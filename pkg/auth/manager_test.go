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

func Test_Auth_002(t *testing.T) {
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
	user, err := manager.CreateUser(context.TODO(), schema.UserMeta{
		Name:  types.StringPtr("test"),
		Desc:  types.StringPtr("test user"),
		Scope: []string{},
		Meta:  map[string]any{},
	})
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create a new user and archive
	archived, err := manager.CreateUser(context.TODO(), schema.UserMeta{
		Name:  types.StringPtr("test2"),
		Desc:  types.StringPtr("test user"),
		Scope: []string{},
		Meta:  map[string]any{},
	})
	if !assert.NoError(err) {
		t.FailNow()
	}
	_, err = manager.DeleteUser(context.TODO(), types.PtrString(archived.Name), false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create a new token
	t.Run("CreateToken1", func(t *testing.T) {
		token, err := manager.CreateToken(context.TODO(), *user.Name, schema.TokenMeta{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotEmpty(token.Id)
		assert.Equal(types.PtrString(user.Name), token.User)
		assert.Equal("live", token.Status)
		assert.NotEmpty(types.PtrString(token.Value))
	})

	// Create a token for an archived user
	t.Run("CreateToken2", func(t *testing.T) {
		_, err := manager.CreateToken(context.TODO(), *archived.Name, schema.TokenMeta{})
		assert.ErrorIs(err, httpresponse.ErrConflict)
	})

	// Create a token then get it
	t.Run("GetToken1", func(t *testing.T) {
		token, err := manager.CreateToken(context.TODO(), *user.Name, schema.TokenMeta{
			Desc: types.StringPtr("test description"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Empty the value field
		token.Value = nil

		token2, err := manager.GetToken(context.TODO(), token.User, token.Id)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(token, token2)
	})

	// Get non-existent token
	t.Run("GetToken2", func(t *testing.T) {
		_, err := manager.GetToken(context.TODO(), "no name", 42)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Create token, archive user, then get token which should be archived
	t.Run("GetToken3", func(t *testing.T) {
		// Create a new user
		user, err := manager.CreateUser(context.TODO(), schema.UserMeta{
			Name:  types.StringPtr("test3"),
			Desc:  types.StringPtr("test user"),
			Scope: []string{},
			Meta:  map[string]any{},
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Create token
		token, err := manager.CreateToken(context.TODO(), *user.Name, schema.TokenMeta{
			Desc: types.StringPtr("test description"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Archive user
		archived, err = manager.DeleteUser(context.TODO(), types.PtrString(user.Name), false)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("archived", archived.Status)

		// Get token - should report as archived
		token2, err := manager.GetToken(context.TODO(), token.User, token.Id)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal("archived", token2.Status)
	})

	// Delete non-existent token
	t.Run("DeleteToken1", func(t *testing.T) {
		_, err := manager.DeleteToken(context.TODO(), "no name", 42, false)
		assert.ErrorIs(err, httpresponse.ErrNotFound)
	})

	// Create a token then archive it
	t.Run("DeleteToken2", func(t *testing.T) {
		// Create token
		token, err := manager.CreateToken(context.TODO(), *user.Name, schema.TokenMeta{
			Desc: types.StringPtr("test description"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Archive the token
		token2, err := manager.DeleteToken(context.TODO(), token.User, token.Id, false)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(token.Id, token2.Id)
		assert.Equal(token.User, token2.User)
		assert.Equal("archived", token2.Status)
	})

	// Create a token then delete it
	t.Run("DeleteToken3", func(t *testing.T) {
		// Create token
		token, err := manager.CreateToken(context.TODO(), *user.Name, schema.TokenMeta{
			Desc: types.StringPtr("test description"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Delete the token
		token2, err := manager.DeleteToken(context.TODO(), token.User, token.Id, true)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(token.Id, token2.Id)
		assert.Equal(token.User, token2.User)
		assert.Equal("deleted", token2.Status)
	})

	// Update a token
	t.Run("UpdateToken", func(t *testing.T) {
		// Create token
		token, err := manager.CreateToken(context.TODO(), *user.Name, schema.TokenMeta{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotEmpty(token.Id)
		assert.Nil(token.Desc)

		// Update token
		token2, err := manager.UpdateToken(context.TODO(), token.User, token.Id, schema.TokenMeta{
			Desc: types.StringPtr("new test description"),
		})
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(token.Id, token2.Id)
		assert.Equal(token.User, token2.User)
		assert.Equal("new test description", types.PtrString(token2.Desc))
	})

	// List tokens for user
	t.Run("ListTokens", func(t *testing.T) {
		// List tokens
		list, err := manager.ListTokens(context.TODO(), *user.Name, schema.TokenListRequest{})
		if !assert.NoError(err) {
			t.FailNow()
		}
		// Default live state
		assert.Equal("live", types.PtrString(list.Status))
	})
}
