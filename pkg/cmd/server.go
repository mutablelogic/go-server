package cmd

import (
	"crypto/tls"
	"fmt"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	openapihttphandler "github.com/mutablelogic/go-server/pkg/openapi/httphandler"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	otel "github.com/mutablelogic/go-server/pkg/otel"
	errgroup "golang.org/x/sync/errgroup"
)

// RegisterFunc is called after the router is created but before the server
// starts listening. Use it to add routes and wire up handlers.
type RegisterFunc func(*httprouter.Router) error

///////////////////////////////////////////////////////////////////////////////
// TYPES

// RunServer is a general-purpose "run" command. Embed it in your CLI's command
// struct to get a fully functional HTTP server with logging and OTel middleware.
type RunServer struct {
	OpenAPI bool `name:"openapi" help:"Serve OpenAPI spec at {prefix}/openapi.{json,yaml,html}" default:"true" negatable:""`

	// TLS server options
	TLS struct {
		ServerName string `name:"name" help:"TLS server name"`
		CertFile   string `name:"cert" help:"TLS certificate file"`
		KeyFile    string `name:"key" help:"TLS key file"`
	} `embed:"" prefix:"tls."`

	// HTTP server options
	HTTP struct {
		Origin string `name:"origin" help:"Cross-origin protection (CSRF) origin. Empty string for same-origin only, '*' to allow all cross-origin requests, or a specific origin in the form 'scheme://host[:port]'." default:""`
	} `embed:"" prefix:"http."`

	register []RegisterFunc
}

// Register appends fns to the list of functions called to wire up routes
// before the server starts. Returns the receiver for chaining.
func (s *RunServer) Register(fns ...RegisterFunc) *RunServer {
	s.register = append(s.register, fns...)
	return s
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (s *RunServer) Run(ctx server.Cmd) error {
	// Build optional server options
	var serverOpts []httpserver.Opt
	if ctx.HTTPTimeout() > 0 {
		serverOpts = append(serverOpts,
			httpserver.WithReadTimeout(ctx.HTTPTimeout()),
			httpserver.WithWriteTimeout(ctx.HTTPTimeout()),
		)
	}

	// Create TLS config optionally when either cert or key file is provided.
	var data [][]byte
	for _, field := range []string{s.TLS.CertFile, s.TLS.KeyFile} {
		if field == "" {
			continue
		}
		if data_, err := os.ReadFile(field); err != nil {
			return fmt.Errorf("%s: %w", field, err)
		} else {
			data = append(data, data_)
		}
	}

	// Set the TLS server name
	var tlsCfg *tls.Config
	if len(data) > 0 {
		var err error
		if tlsCfg, err = httpserver.TLSConfig(s.TLS.ServerName, false, data...); err != nil {
			return fmt.Errorf("tls: %w", err)
		}
	} else if s.TLS.ServerName != "" {
		tlsCfg = &tls.Config{ServerName: s.TLS.ServerName}
	}

	// Create a new server and start listening. The server will run until the context is cancelled.
	srv, err := httpserver.New(ctx.HTTPAddr(), tlsCfg, serverOpts...)
	if err != nil {
		return fmt.Errorf("httpserver: %w", err)
	}

	// Build middleware chain: OTel HTTP middleware which emits traces, logs and metrics
	middleware := []httprouter.HTTPMiddlewareFunc{
		otel.HTTPHandlerFunc(srv.URL().Host, ctx.Logger()),
	}

	// Create the router
	router, err := httprouter.NewRouter(ctx.Context(), srv.Router(), ctx.HTTPPrefix(), s.HTTP.Origin, ctx.Name(), ctx.Version(), middleware...)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	// Register routes
	for _, fn := range s.register {
		if err := fn(router); err != nil {
			return fmt.Errorf("register: %w", err)
		}
	}

	// Register OpenAPI spec endpoints if enabled
	if s.OpenAPI {
		if err := openapihttphandler.RegisterHandler(router, false); err != nil {
			return fmt.Errorf("openapi: %w", err)
		}
	}

	// Always register a catch-all 404 handler at "/" and at the prefix root (e.g. "/api")
	if err := router.RegisterCatchAll("/", false); err != nil {
		return fmt.Errorf("catchall: %w", err)
	}
	if err := router.RegisterCatchAll(ctx.HTTPPrefix(), false); err != nil {
		return fmt.Errorf("catchall: %w", err)
	}

	// Bind to the server's address to ensure it's available
	if err := srv.Listen(); err != nil {
		return err
	}

	// Set the server URL in the OpenAPI spec now that the bound address is known.
	// When a TLS server name is configured (e.g. behind a reverse proxy) it is
	// used as the public hostname; otherwise the bound listen address is used.
	if s.OpenAPI {
		router.Spec().SetSummary(ctx.Description())
		router.Spec().SetServers([]openapi.Server{
			{URL: srv.URL().String()},
		})
	}

	// Report the server endpoint and version
	ctx.Logger().InfoContext(ctx.Context(), "started", "name", ctx.Name(), "version", ctx.Version(), "addr", srv.Addr(), "server", srv.URL().String(), "origin", s.HTTP.Origin, "prefix", ctx.HTTPPrefix())

	// Run the server until the context is cancelled
	eg, egCtx := errgroup.WithContext(ctx.Context())
	eg.Go(func() error {
		return srv.Run(egCtx)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	// Return success
	ctx.Logger().InfoContext(ctx.Context(), "terminated gracefully")
	return nil
}
