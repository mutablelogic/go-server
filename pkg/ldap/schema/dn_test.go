package schema_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_DN_001(t *testing.T) {
	assert := assert.New(t)

	// Check new DN
	dn, err := schema.NewDN("cn=John Doe,dc=example,dc=com")
	assert.NoError(err)
	assert.NotNil(dn)
	assert.Equal("cn=John Doe,dc=example,dc=com", dn.String())
}

func Test_DN_002(t *testing.T) {
	assert := assert.New(t)

	// Setup
	bdn, err := schema.NewDN("ou=users,dc=example,dc=com")
	assert.NoError(err)

	rdn, err := schema.NewDN("cn=John Doe")
	assert.NoError(err)

	dn := rdn.Join(bdn)
	assert.NotNil(dn)
	assert.Equal("cn=John Doe,ou=users,dc=example,dc=com", dn.String())

	// Check ancestor
	assert.True(bdn.AncestorOf(dn))

}
