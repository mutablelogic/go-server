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
	v := ctx.Version()

	// Build middleware chain: logger always first, OTel if configured
	middleware := []httprouter.HTTPMiddlewareFunc{logger.NewMiddleware(ctx.Logger())}
	if ctx.Tracer() != nil {
		middleware = append(middleware, otel.HTTPHandlerFunc(ctx.Tracer()))
	}

	// Create the router
	router, err := httprouter.NewRouter(ctx.Context(), ctx.HTTPPrefix(), ctx.HTTPOrigin(), ctx.Name(), v, middleware...)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	// Build optional server options
	var serverOpts []httpserver.Opt
	if ctx.HTTPTimeout() > 0 {
		serverOpts = append(serverOpts,
			httpserver.WithReadTimeout(ctx.HTTPTimeout()),
			httpserver.WithWriteTimeout(ctx.HTTPTimeout()),
		)
	}

	// Create TLS config optionally when either cert or key file is provided.
	var tlsCfg *tls.Config
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
	if len(data) > 0 {
		if tlsCfg, err = httpserver.TLSConfig(s.TLS.ServerName, false, data...); err != nil {
			return fmt.Errorf("tls: %w", err)
		}
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

	// Always register a catch-all 404 handler at "/"
	if err := router.RegisterCatchAll(false); err != nil {
		return fmt.Errorf("catchall: %w", err)
	}

	// Create a new server and start listening. The server will run until the context is cancelled.
	srv, err := httpserver.New(ctx.HTTPAddr(), router, tlsCfg, serverOpts...)
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
