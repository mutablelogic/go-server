package main

import (
	"context"
	"fmt"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Plugins
	auth "github.com/mutablelogic/go-server/pkg/auth/config"
	cert "github.com/mutablelogic/go-server/pkg/cert/config"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter/config"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver/config"
	logger "github.com/mutablelogic/go-server/pkg/logger/config"
	pg "github.com/mutablelogic/go-server/pkg/pgmanager/config"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue/config"

	// Static content
	helloworld "github.com/mutablelogic/go-server/npm/helloworld"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServiceCommands struct {
	Run ServiceRunCommand `cmd:"" group:"SERVICE" help:"Run the service"`
}

type ServiceRunCommand struct {
	Plugins []string `help:"Plugin paths"`
	Router  struct {
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
	PGQueue struct {
		pgqueue.Config `embed:"" prefix:"pgqueue."` // Postgresql queue configuration
	} `embed:""`
	CertManager struct {
		cert.Config `embed:"" prefix:"cert."` // Certificate manager configuration
	} `embed:""`
	Log struct {
		logger.Config `embed:"" prefix:"log."` // Logger configuration
	} `embed:""`
	HelloWorld struct {
		helloworld.Config `embed:"" prefix:"helloworld."` // HelloWorld configuration
	} `embed:""`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ServiceRunCommand) Run(app server.Cmd) error {
	// Load plugins
	plugins, err := provider.LoadPluginsForPattern(cmd.Plugins...)
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		fmt.Println("TODO: Loaded plugins:", plugin.Name())
	}

	// Set the server listener and router prefix
	cmd.Server.Listen = app.GetEndpoint()
	cmd.Router.Prefix = types.NormalisePath(cmd.Server.Listen.Path)

	// Create a provider and resolve references
	provider, err := provider.New(func(ctx context.Context, label string, plugin server.Plugin) (server.Plugin, error) {
		ref.Log(ctx).Debugf(ctx, "Resolving %q", label)
		switch label {
		case "log":
			config := plugin.(logger.Config)
			config.Debug = app.GetDebug() >= server.Debug
			return config, nil

		case "certmanager":
			config := plugin.(cert.Config)

			// Set the router
			if router, ok := ref.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", "httprouter")
			} else {
				config.Router = router
			}

			// Set the connection pool
			if pool, ok := ref.Provider(ctx).Task(ctx, "pgpool").(server.PG); !ok || pool == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid connection pool %q", "pgpool")
			} else {
				config.Pool = pool
			}

			// Set the queue
			if queue, ok := ref.Provider(ctx).Task(ctx, "pgqueue").(server.PGQueue); !ok || queue == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid connection pool %q", "pgqueue")
			} else {
				config.Queue = queue
			}

			return config, nil

		case "httpserver":
			config := plugin.(httpserver.Config)

			// Set the router
			if router, ok := ref.Provider(ctx).Task(ctx, "httprouter").(http.Handler); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", "httprouter")
			} else {
				config.Router = router
			}

			// Return the new configuration with the router
			return config, nil

		case "httprouter":
			config := plugin.(httprouter.Config)

			// Set the middleware
			config.Middleware = []string{}

			// Return the new configuration with the router
			return config, nil

		case "helloworld":
			config := plugin.(helloworld.Config)

			// Set the router
			if router, ok := ref.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", "httprouter")
			} else {
				config.Router = router
			}

			// Return the new configuration with the router
			return config, nil

		case "auth":
			config := plugin.(auth.Config)

			// Set the router
			if router, ok := ref.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", "httprouter")
			} else {
				config.Router = router
			}

			// Set the connection pool
			if pool, ok := ref.Provider(ctx).Task(ctx, "pgpool").(server.PG); !ok || pool == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid connection pool %q", "pgpool")
			} else {
				config.Pool = pool
			}

			// Return the new configuration
			return config, nil

		case "pgqueue":
			config := plugin.(pgqueue.Config)

			// Set the router
			if router, ok := ref.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", "httprouter")
			} else {
				config.Router = router
			}

			// Set the connection pool
			if pool, ok := ref.Provider(ctx).Task(ctx, "pgpool").(server.PG); !ok || pool == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid connection pool %q", "pgpool")
			} else {
				config.Pool = pool
			}

			// Return the new configuration
			return config, nil

		case "pgpool":
			config := plugin.(pg.Config)

			// Set the router
			if router, ok := ref.Provider(ctx).Task(ctx, "httprouter").(server.HTTPRouter); !ok || router == nil {
				return nil, httpresponse.ErrInternalError.Withf("Invalid router %q", "httprouter")
			} else {
				config.Router = router
			}

			// Set trace
			if app.GetDebug() == server.Trace {
				config.Trace = func(ctx context.Context, query string, args any, err error) {
					if err != nil {
						ref.Log(ctx).With("args", args).Print(ctx, err, " ON ", query)
					} else {
						ref.Log(ctx).With("args", args).Debug(ctx, query)
					}
				}
			}

			// Return the new configuration with the router
			return config, nil
		}

		// No-op
		return plugin, nil
	}, cmd.Log.Config, cmd.Router.Config, cmd.Server.Config, cmd.HelloWorld.Config, cmd.Auth.Config, cmd.PGPool.Config, cmd.PGQueue.Config, cmd.CertManager.Config)
	if err != nil {
		return err
	}

	// Run the provider
	return provider.Run(app.Context())
}
