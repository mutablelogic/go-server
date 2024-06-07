package schema

import (
	"fmt"
	"slices"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Schema is the schema for the LDAP server
type Schema struct {
	DN               string
	UserOU           string
	GroupOU          string
	UserObjectClass  []string
	GroupObjectClass []string
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultPosixUserObjectClass = "posixAccount"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s Schema) IsPosix() bool {
	return slices.Contains(s.UserObjectClass, defaultPosixUserObjectClass)
}

func (s Schema) GroupDN(name string) string {
	return fmt.Sprintf("cn=%s,ou=%s,%s", name, s.GroupOU, s.DN)
}

func (s Schema) NewObject(entry *goldap.Entry) *Object {
	return &Object{Entry: entry}
}

func (s Schema) NewGroup(name string) *Object {
	// Check parameters
	if !types.IsIdentifier(name) || s.GroupOU == "" {
		return nil
	}
	group := &Object{Entry: &goldap.Entry{
		DN: s.GroupDN(name),
	}}
	group.Set("objectClass", s.GroupObjectClass...)
	return group
}
