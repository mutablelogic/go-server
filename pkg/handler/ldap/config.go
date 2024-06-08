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
	Schema schema.Schema `hcl:"schema,optional" description:"LDAP Schema"`
}

// Ensure that LDAP implements the Service interface
var _ server.Plugin = (*Config)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName         = "ldap"
	defaultMethodPlain  = "ldap"
	defaultMethodSecure = "ldaps"
	defaultPortPlain    = 389
	defaultPortSecure   = 636
	deltaPingTime       = time.Minute
	defaultUserOU       = "users"
	defaultGroupOU      = "groups"
)

var (
	defaultUserObjectClass  = []string{"posixAccount", "top", "person", "inetOrgPerson"}
	defaultGroupObjectClass = []string{"posixGroup", "top", "groupOfUniqueNames"}
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

func (c Config) ObjectSchema() *schema.Schema {
	schema := &schema.Schema{
		UserObjectClass:  defaultUserObjectClass,
		GroupObjectClass: defaultGroupObjectClass,
		UserOU:           defaultUserOU,
		GroupOU:          defaultGroupOU,
	}

	if len(c.Schema.UserObjectClass) > 0 {
		schema.UserObjectClass = c.Schema.UserObjectClass
	}
	if len(c.Schema.GroupObjectClass) > 0 {
		schema.GroupObjectClass = c.Schema.GroupObjectClass
	}
	if c.Schema.UserOU != "" && c.Schema.GroupOU != "" {
		schema.UserOU = c.Schema.UserOU
		schema.GroupOU = c.Schema.GroupOU
	}
	return schema
}
