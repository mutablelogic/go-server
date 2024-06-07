package ldap

import (
	// Packages
	"time"

	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/handler/ldap/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// LDAP configuration
type Config struct {
	URL      string `hcl:"url,required" description:"LDAP server URL, user and optional password for bind"`
	User     string `hcl:"user,optional" description:"Bind user, if not specified in URL"`
	Password string `hcl:"password,optional" description:"Bind password, if not specified in URL"`
	DN       string `hcl:"dn,required" description:"Distinguished name"`
	TLS      struct {
		SkipVerify bool `hcl:"skip-verify" description:"Skip certificate verification"`
	} `hcl:"tls,optional" description:"TLS configuration"`
	Schema struct {
		User  []string `hcl:"user,optional" description:"User objectClass, defaults to posixAccount, person, inetOrgPerson"`
		Group []string `hcl:"group,optional" description:"Group objectClass, defaults to posixGroup"`
	} `hcl:"schema,optional" description:"LDAP Schema"`
}

// Ensure that LDAP implements the Service interface
var _ server.Plugin = (*Config)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName         = "ldap"
	defaultMethodPlain  = "ldap"
	defaultPortPlain    = 389
	defaultMethodSecure = "ldaps"
	defaultPortSecure   = 636
	deltaPingTime       = time.Minute
	defaultUserOU       = "users"
	defaultGroupOU      = "groups"
)

var (
	defaultUserObjectClass  = []string{"posixAccount", "person", "inetOrgPerson"}
	defaultGroupObjectClass = []string{"posixGroup"}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) Name() string {
	return defaultName
}

func (c Config) Description() string {
	return "Provides a client for LDAP communication"
}

func (c Config) New() (server.Task, error) {
	return New(c)
}

func (c Config) ObjectSchema() (*schema.Schema, error) {
	schema := &schema.Schema{
		DN:               c.DN,
		UserObjectClass:  defaultUserObjectClass,
		GroupObjectClass: defaultGroupObjectClass,
		UserOU:           defaultUserOU,
		GroupOU:          defaultGroupOU,
	}
	if len(c.Schema.User) > 0 {
		schema.UserObjectClass = c.Schema.User
	}
	if len(c.Schema.Group) > 0 {
		schema.GroupObjectClass = c.Schema.Group
	}
	return schema, nil
}
