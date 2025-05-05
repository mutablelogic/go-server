package schema

import (
	"strings"
	"time"
)

const (
	SchemaName   = "ldap"
	APIPrefix    = "/ldap/v1"
	MethodPlain  = "ldap"
	MethodSecure = "ldaps"
	PortPlain    = 389
	PortSecure   = 636

	// Time between connection retries
	MinRetryInterval = time.Second * 5
	MaxRetries       = 10

	// Maximum number of entries to return in a single request
	MaxListPaging = 500

	// Maximum list entries to return
	MaxListEntries = 1000

	// Attributes
	AttrObjectClasses  = "objectClasses"
	AttrAttributeTypes = "attributeTypes"
	AttrSubSchemaDN    = "subschemaSubentry"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type DN string

type Group struct {
	DN          DN       `json:"dn,omitempty"`
	ObjectClass []string `json:"objectclass,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d DN) Join(parts ...string) DN {
	return DN(strings.Join(append([]string{string(d)}, parts...), ","))
}
