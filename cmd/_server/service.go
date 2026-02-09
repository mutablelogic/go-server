package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	helloworld "github.com/mutablelogic/go-server/npm/helloworld"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"
	version "github.com/mutablelogic/go-server/pkg/version"

	// Plugins
	auth "github.com/mutablelogic/go-server/pkg/auth/config"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter/config"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver/config"
	ldap "github.com/mutablelogic/go-server/pkg/ldap/config"
	logger "github.com/mutablelogic/go-server/pkg/logger/config"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServiceCommands struct {
	//	Run  ServiceRunCommand  `cmd:"" group:"SERVICE" help:"Run the service"`
	Run    ServiceRunCommand    `cmd:"" group:"SERVICE" help:"Run the service with plugins"`
	Config ServiceConfigCommand `cmd:"" group:"SERVICE" help:"Output the plugin configuration"`
}

type ServiceRunCommand struct {
	Plugins []string `help:"Plugin paths" env:"PLUGIN_PATH"`
}

type ServiceConfigCommand struct {
	ServiceRunCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ServiceConfigCommand) Run(app server.Cmd) error {
	// Create a provider by loading the plugins
	provider, err := provider.NewWithPlugins(cmd.Plugins...)
	if err != nil {
		return err
	}
	return provider.WriteConfig(os.Stdout)
}

func (cmd *ServiceRunCommand) Run(app server.Cmd) error {
	// Create a provider by loading the plugins
	provider, err := provider.NewWithPlugins(cmd.Plugins...)
	if err != nil {
		return err
	}

	endpoint, err := app.GetEndpoint()
	if err != nil {
		return err
	}

	// Set the configuration
	err = errors.Join(err, provider.Load("log", "main", func(ctx context.Context, label string, config server.Plugin) error {
		logger := config.(*logger.Config)
		logger.Debug = app.GetDebug() >= server.Debug
		return nil
	}))
	err = errors.Join(err, provider.Load("httprouter", "main", func(ctx context.Context, label string, config server.Plugin) error {
		httprouter := config.(*httprouter.Config)
		httprouter.Prefix = types.NormalisePath(endpoint.Path)
		httprouter.Origin = "*"
		httprouter.Middleware = []string{"log.main"}
		return nil
	}))
	err = errors.Join(err, provider.Load("httpserver", "main", func(ctx context.Context, label string, config server.Plugin) error {
		httpserver := config.(*httpserver.Config)
		httpserver.Listen = endpoint

		// Set router
		if router, ok := provider.Task(ctx, "httprouter.main").(http.Handler); !ok || router == nil {
			return httpresponse.ErrInternalError.With("Invalid router")
		} else {
			httpserver.Router = router
		}

		// Return success
		return nil
	}))
	err = errors.Join(err, provider.Load("helloworld", "main", func(ctx context.Context, label string, config server.Plugin) error {
		helloworld := config.(*helloworld.Config)

		// Set router
		if router, ok := provider.Task(ctx, "httprouter.main").(server.HTTPRouter); !ok || router == nil {
			return httpresponse.ErrInternalError.With("Invalid router")
		} else {
			helloworld.Router = router
		}

		// Return success
		return nil
	}))

	err = errors.Join(err, provider.Load("auth", "main", func(ctx context.Context, label string, config server.Plugin) error {
		auth := config.(*auth.Config)

		// Set the router
		if router, ok := ref.Provider(ctx).Task(ctx, "httprouter.main").(server.HTTPRouter); !ok || router == nil {
			return httpresponse.ErrInternalError.With("Invalid router")
		} else {
			auth.Router = router
		}

		// Set the connection pool
		if pool, ok := ref.Provider(ctx).Task(ctx, "pgpool.main").(server.PG); !ok || pool == nil {
			return httpresponse.ErrInternalError.With("Invalid connection pool")
		} else {
			auth.Pool = pool
		}

		// Return success
		return nil
	}))

	err = errors.Join(err, provider.Load("ldap", "main", func(ctx context.Context, label string, config server.Plugin) error {
		ldap := config.(*ldap.Config)

		// Set the router
		if router, ok := ref.Provider(ctx).Task(ctx, "httprouter.main").(server.HTTPRouter); !ok || router == nil {
			return httpresponse.ErrInternalError.With("Invalid router")
		} else {
			ldap.Router = router
		}

		// HACK
		ldap.UserSchema.RDN = "cn=users,cn=accounts"
		ldap.UserSchema.Field = "uid"
		ldap.UserSchema.ObjectClasses = "top,inetOrgPerson,person,posixAccount"

		ldap.GroupSchema.RDN = "cn=groups,cn=accounts"
		ldap.GroupSchema.Field = "cn"
		ldap.GroupSchema.ObjectClasses = "top,groupOfNames,nestedGroup,posixGroup"

		return nil
	}))

	if err != nil {
		return err
	}

	fmt.Println("go-server", version.Version())

	// Run the provider
	return provider.Run(app.Context())
}
