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

}
