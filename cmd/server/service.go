package main

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Plugins
	auth "github.com/mutablelogic/go-server/plugin/auth"
	certmanager "github.com/mutablelogic/go-server/plugin/certmanager"
	httprouter "github.com/mutablelogic/go-server/plugin/httprouter"
	httpserver "github.com/mutablelogic/go-server/plugin/httpserver"
	log "github.com/mutablelogic/go-server/plugin/log"
	pg "github.com/mutablelogic/go-server/plugin/pg"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServiceCommands struct {
	Run ServiceRunCommand `cmd:"" group:"SERVICE" help:"Run the service"`
}

type ServiceRunCommand struct {
	Router struct {
		httprouter.Config `embed:"" prefix:"router."` // Router configuration
	} `embed:""`
	Server struct {
		httpserver.Config `embed:"" prefix:"server."` // Server configuration
	} `embed:""`
	Auth struct {
		auth.Config `embed:"" prefix:"auth."` // Auth configuration
	} `embed:""`
	PGPool struct {
		pg.Config `embed:"" prefix:"pg."` // Postgresql configuration
	} `embed:""`
	CertManager struct {
		certmanager.Config `embed:"" prefix:"cert."` // Certificate manager configuration
	} `embed:""`
	Log struct {
		log.Config `embed:"" prefix:"log."` // Logger configuration
	} `embed:""`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ServiceRunCommand) Run(app server.Cmd) error {
	// Set the server listener and router prefix
	cmd.Server.Listen = app.GetEndpoint()
	cmd.Router.Prefix = types.NormalisePath(cmd.Server.Listen.Path)

	// Create a provider and resolve references
	provider, err := provider.New(func(ctx context.Context, label string, plugin server.Plugin) (server.Plugin, error) {
		provider.Log(ctx).Debugf(ctx, "Resolving %q", label)
		switch label {
		case "log":
			config := plugin.(log.Config)
			config.Debug = app.GetDebug() >= server.Debug
			return config, nil
		case "certmanager":
			config := plugin.(certmanager.Config)

			// Set the router
			if router, ok := provider.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", label)
			} else {
				config.Router = router
			}

			// Set the connection pool
			if pool, ok := provider.Provider(ctx).Task(ctx, "pgpool").(server.PG); !ok || pool == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid connection pool %q", label)
			} else {
				config.Pool = pool
			}

			return config, nil

		case "httpserver":
			config := plugin.(httpserver.Config)

			// Set the router
			if router, ok := provider.Provider(ctx).Task(ctx, "httprouter").(http.Handler); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", label)
			} else {
				config.Router = router
			}

			// Return the new configuration with the router
			return config, nil

		case "httprouter":
			config := plugin.(httprouter.Config)

			// Set the middleware
			config.Middleware = []string{"log"}

			// Return the new configuration with the router
			return config, nil

		case "auth":
			config := plugin.(auth.Config)

			// Set the router
			if router, ok := provider.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", label)
			} else {
				config.Router = router
			}

			// Set the connection pool
			if pool, ok := provider.Provider(ctx).Task(ctx, "pgpool").(server.PG); !ok || pool == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid connection pool %q", label)
			} else {
				config.Pool = pool
			}

			// Return the new configuration
			return config, nil

		case "pgpool":
			config := plugin.(pg.Config)

			// Set the router
			if router, ok := provider.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", label)
			} else {
				config.Router = router
			}

			// Set trace
			if app.GetDebug() == server.Trace {
				config.Trace = func(ctx context.Context, query string, args any, err error) {
					if err != nil {
						provider.Log(ctx).With("args", args).Print(ctx, err, " ON ", query)
					} else {
						provider.Log(ctx).With("args", args).Debug(ctx, query)
					}
				}
			}

			// Return the new configuration with the router
			return config, nil
		}

		// No-op
		return plugin, nil
	}, cmd.Log.Config, cmd.Router.Config, cmd.Server.Config, cmd.Auth.Config, cmd.PGPool.Config, cmd.CertManager.Config)
	if err != nil {
		return err
	}

	// Run the provider
	return provider.Run(app.Context())
}
