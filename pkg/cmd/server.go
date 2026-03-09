package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	openapihttphandler "github.com/mutablelogic/go-server/pkg/openapi/httphandler"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	otel "github.com/mutablelogic/go-server/pkg/otel"
	errgroup "golang.org/x/sync/errgroup"
)

// RegisterFunc is called after the router is created but before the server
// starts listening. Use it to add routes and wire up handlers.
type RegisterFunc func(*httprouter.Router, server.Cmd) error

///////////////////////////////////////////////////////////////////////////////
// TYPES

// RunServer is a general-purpose "run" command. Embed it in your CLI's command
// struct to get a fully functional HTTP server with logging and OTel middleware.
type RunServer struct {
	OpenAPI bool `name:"openapi" help:"Serve OpenAPI spec at {prefix}/openapi.json" default:"true" negatable:""`

	// TLS server options
	TLS struct {
		ServerName string `name:"name" help:"TLS server name"`
		CertFile   string `name:"cert" help:"TLS certificate file"`
		KeyFile    string `name:"key" help:"TLS key file"`
	} `embed:"" prefix:"tls."`

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

func (s *RunServer) Run(ctx *Global) error {
	v := ctx.Version()

	// Build middleware chain: logger always first, OTel if configured
	middleware := []httprouter.HTTPMiddlewareFunc{logger.NewMiddleware(ctx.Logger())}
	if ctx.tracer != nil {
		middleware = append(middleware, otel.HTTPHandlerFunc(ctx.tracer))
	}

	// Create the router
	router, err := httprouter.NewRouter(ctx.Context(), ctx.HTTP.Prefix, ctx.HTTP.Origin, ctx.Name(), v, middleware...)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	// Build optional server options
	var serverOpts []httpserver.Opt
	if ctx.HTTP.Timeout > 0 {
		serverOpts = append(serverOpts,
			httpserver.WithReadTimeout(ctx.HTTP.Timeout),
			httpserver.WithWriteTimeout(ctx.HTTP.Timeout),
		)
	}

	var tlsCfg *tls.Config
	if s.TLS.CertFile != "" || s.TLS.KeyFile != "" {
		certPEM, err := os.ReadFile(s.TLS.CertFile)
		if err != nil {
			return fmt.Errorf("tls.cert: %w", err)
		}
		keyPEM, err := os.ReadFile(s.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("tls.key: %w", err)
		}
		if tlsCfg, err = httpserver.TLSConfig(s.TLS.ServerName, false, certPEM, keyPEM); err != nil {
			return fmt.Errorf("tls: %w", err)
		}
	}

	// Register routes
	for _, fn := range s.register {
		if err := fn(router, ctx); err != nil {
			return fmt.Errorf("register: %w", err)
		}
	}

	// Register OpenAPI spec endpoints if enabled
	if s.OpenAPI {
		if err := openapihttphandler.RegisterHandler(router, false); err != nil {
			return fmt.Errorf("openapi: %w", err)
		}
	}

	// Always register a catch-all 404 handler at "/"
	if err := router.RegisterCatchAll(false); err != nil {
		return fmt.Errorf("catchall: %w", err)
	}

	// Create a new server and start listening. The server will run until the context is cancelled.
	srv, err := httpserver.New(ctx.HTTP.Addr, router, tlsCfg, serverOpts...)
	if err != nil {
		return fmt.Errorf("httpserver: %w", err)
	}

	if err := srv.Listen(); err != nil {
		return err
	}

	// Set the server URL in the OpenAPI spec now that the bound address is known
	if s.OpenAPI {
		scheme := "http"
		if tlsCfg != nil {
			scheme = "https"
		}
		router.Spec().SetSummary(ctx.Description())
		router.Spec().SetServers([]openapi.Server{
			{URL: fmt.Sprintf("%s://%s", scheme, srv.Addr())},
		})
	}

	// Report the server endpoint and version
	ctx.Logger().InfoContext(ctx.Context(), "started", "name", ctx.Name(), "version", v, "endpoint", srv.Addr())

	// Run the server until the context is cancelled
	eg, egCtx := errgroup.WithContext(ctx.Context())
	eg.Go(func() error {
		return srv.Run(egCtx)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	// Return success
	ctx.Logger().InfoContext(context.Background(), "terminated gracefully")
	return nil
}
