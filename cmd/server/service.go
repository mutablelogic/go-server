package main

import (
	"context"
	"errors"
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
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServiceCommands struct {
	Run  ServiceRunCommand  `cmd:"" group:"SERVICE" help:"Run the service"`
	Run2 ServiceRun2Command `cmd:"" group:"SERVICE" help:"Run the service with plugins"`
}

type ServiceRun2Command struct {
	Plugins []string `help:"Plugin paths"`
}

type ServiceRunCommand struct {
	Router struct {
		httprouter.Config `embed:"" prefix:"router."` // Router configuration
	} `embed:"" prefix:""`
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
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ServiceRunCommand) Run(app server.Cmd) error {
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
	}, cmd.Log.Config, cmd.Router.Config, cmd.Server.Config, cmd.Auth.Config, cmd.PGPool.Config, cmd.PGQueue.Config, cmd.CertManager.Config)
	if err != nil {
		return err
	}

	// Run the provider
	return provider.Run(app.Context())
}

func (cmd *ServiceRun2Command) Run(app server.Cmd) error {
	// Create a provider by loading the plugins
	provider, err := provider.NewWithPlugins(cmd.Plugins...)
	if err != nil {
		return err
	}

	// Set the configuration
	err = errors.Join(err, provider.Load("log", "main", func(ctx context.Context, config server.Plugin) {
		logger := config.(*logger.Config)
		logger.Debug = app.GetDebug() >= server.Debug
	}))
	err = errors.Join(err, provider.Load("httprouter", "main", func(ctx context.Context, config server.Plugin) {
		httprouter := config.(*httprouter.Config)
		httprouter.Prefix = types.NormalisePath(app.GetEndpoint().Path)
		httprouter.Origin = "*"
		httprouter.Middleware = []string{"log.main"}
	}))
	err = errors.Join(err, provider.Load("httpserver", "main", func(ctx context.Context, config server.Plugin) {
		httpserver := config.(*httpserver.Config)
		httpserver.Listen = app.GetEndpoint()
		httpserver.Router = provider.Task(ctx, "httprouter.main").(http.Handler)
	}))
	err = errors.Join(err, provider.Load("helloworld", "main", func(ctx context.Context, config server.Plugin) {
		// NO-OP
	}))
	err = errors.Join(err, provider.Load("pgpool", "main", func(ctx context.Context, config server.Plugin) {
		pgpool := config.(*pg.Config)
		pgpool.Router = provider.Task(ctx, "httprouter.main").(server.HTTPRouter)
	}))
	err = errors.Join(err, provider.Load("auth", "main", func(ctx context.Context, config server.Plugin) {
		auth := config.(*auth.Config)
		auth.Pool = provider.Task(ctx, "pgpool.main").(server.PG)
		auth.Router = provider.Task(ctx, "httprouter.main").(server.HTTPRouter)
	}))
	err = errors.Join(err, provider.Load("pgqueue", "main", func(ctx context.Context, config server.Plugin) {
		pgqueue := config.(*pgqueue.Config)
		pgqueue.Pool = provider.Task(ctx, "pgpool.main").(server.PG)
		pgqueue.Router = provider.Task(ctx, "httprouter.main").(server.HTTPRouter)
	}))
	err = errors.Join(err, provider.Load("certmanager", "main", func(ctx context.Context, config server.Plugin) {
		certmanager := config.(*cert.Config)
		certmanager.Router = provider.Task(ctx, "httprouter.main").(server.HTTPRouter)
		certmanager.Pool = provider.Task(ctx, "pgpool.main").(server.PG)
	}))
	if err != nil {
		return err
	}

	// Run the provider
	return provider.Run(app.Context())
}
