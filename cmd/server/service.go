package main

import (
	"context"
	"log"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Plugins
	certmanager "github.com/mutablelogic/go-server/plugin/certmanager"
	httprouter "github.com/mutablelogic/go-server/plugin/httprouter"
	httpserver "github.com/mutablelogic/go-server/plugin/httpserver"
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
	PGPool struct {
		pg.Config `embed:"" prefix:"pg."` // Postgresql configuration
	} `embed:""`
	CertManager struct {
		certmanager.Config `embed:"" prefix:"cert."` // Certificate manager configuration
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
		switch label {
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
		case "pgpool":
			config := plugin.(pg.Config)

			// Set trace
			if app.GetDebug() {
				config.Trace = func(ctx context.Context, query string, args any, err error) {
					log.Println("PG", query, args, err)
				}
			}

			// Return the new configuration with the router
			return config, nil
		}

		// No-op
		return plugin, nil
	}, cmd.Router.Config, cmd.Server.Config, cmd.PGPool.Config, cmd.CertManager.Config)
	if err != nil {
		return err
	}

	// Run the provider
	return provider.Run(app.Context())
}
