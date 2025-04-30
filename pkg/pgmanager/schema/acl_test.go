package schema_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/pgmanager/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_ACLItem_002(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		acl, err := schema.NewACLItem("miriam=arwdDxtm/miriam")
		if assert.NoError(err) {
			assert.Equal("miriam", acl.Role)
			assert.Equal([]string{"INSERT", "SELECT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER", "MAINTAIN"}, acl.Priv)
			assert.Equal("miriam", acl.Grantor)
			t.Log(acl)
		}
	})

	t.Run("2", func(t *testing.T) {
		acl, err := schema.NewACLItem("=r/miriam")
		if assert.NoError(err) {
			assert.Equal(schema.DefaultAclRole, acl.Role)
			assert.Equal([]string{"SELECT"}, acl.Priv)
			assert.Equal("miriam", acl.Grantor)
			t.Log(acl)
		}
	})

	t.Run("3", func(t *testing.T) {
		acl, err := schema.NewACLItem("miriam=r*w/")
		if assert.NoError(err) {
			assert.Equal("miriam", acl.Role)
			assert.Equal([]string{"SELECT WITH GRANT OPTION", "UPDATE"}, acl.Priv)
			assert.Equal(schema.DefaultAclRole, acl.Grantor)
			t.Log(acl)
		}
	})
}
