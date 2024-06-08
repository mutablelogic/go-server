package schema_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/ldap/schema"
	"github.com/stretchr/testify/assert"
)

func Test_object_001(t *testing.T) {
	assert := assert.New(t)
	o := schema.NewObject("dn")
	assert.NotNil(o)

	o.Set("objectClass", "class1", "class2")
	assert.Equal("dn", o.DN)
	assert.Equal("class1", o.Get("objectClass"))
}
