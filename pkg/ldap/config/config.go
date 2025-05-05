package config

import (
	"context"
	"net/url"

	// Packages
	server "github.com/mutablelogic/go-server"
	ldap "github.com/mutablelogic/go-server/pkg/ldap"
	handler "github.com/mutablelogic/go-server/pkg/ldap/handler"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Url        *url.URL          `env:"LDAP_URL" help:"LDAP connection URL"`                // Connection URL
	User       string            `env:"LDAP_USER" help:"User"`                              // User
	Password   string            `env:"LDAP_PASSWORD" help:"Password"`                      // Password
	BaseDN     string            `env:"LDAP_BASE_DN" help:"Base DN"`                        // Base DN
	SkipVerify bool              `env:"LDAP_SKIPVERIFY" help:"Skip TLS certificate verify"` // Skip verify
	Router     server.HTTPRouter `kong:"-"`                                                 // HTTP Router
}

var _ server.Plugin = Config{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	// Add options
	opts := []ldap.Opt{
		ldap.WithUser(c.User),
		ldap.WithPassword(c.Password),
		ldap.WithBaseDN(c.BaseDN),
	}
	if c.Url != nil {
		opts = append(opts, ldap.WithUrl(c.Url.String()))
	}
	if c.BaseDN != "" {
		opts = append(opts, ldap.WithBaseDN(c.BaseDN))
	}
	if c.SkipVerify {
		opts = append(opts, ldap.WithSkipVerify())
	}

	// Create a new LDAP manager
	manager, err := ldap.NewManager(opts...)
	if err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		handler.Register(ctx, c.Router, manager)
	}

	// Return the task
	return NewTask(manager)
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return schema.SchemaName
}

func (c Config) Description() string {
	return "LDAP manager"
}
