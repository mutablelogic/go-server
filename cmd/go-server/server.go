//go:build !client

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	server "github.com/mutablelogic/go-server"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	httpserver_resource "github.com/mutablelogic/go-server/pkg/httpserver/resource"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	httphandler "github.com/mutablelogic/go-server/pkg/provider/httphandler"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run server." group:"SERVER"`
}

type RunServer struct {
	// TLS server options
	TLS struct {
		ServerName string `name:"name" help:"TLS server name"`
		CertFile   string `name:"cert" help:"TLS certificate file"`
		KeyFile    string `name:"key" help:"TLS key file"`
	} `embed:"" prefix:"tls."`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *RunServer) Run(ctx *Globals) error {
	var result error

	// Create the  manager
	manager, err := provider.New("go-server", "A generic server for running providers", version.GitTag)
	if err != nil {
		return err
	}
	defer func() {
		if err := manager.Close(ctx.ctx); err != nil {
			ctx.logger.Print(ctx.ctx, "error closing manager: ", err)
		}
	}()

	// Register providers with the manager here
	manager.RegisterResource(httpserver_resource.Resource{})

	// Middleware
	middleware := []httprouter.HTTPMiddlewareFunc{}

	// Add logging middleware if the logger implements the Middleware interface
	if logger, ok := ctx.logger.(server.Middleware); ok {
		middleware = append(middleware, logger.HTTPHandlerFunc)
	}

	// If we have an OTEL tracer, add tracing middleware
	if ctx.tracer != nil {
		middleware = append(middleware, otel.HTTPHandlerFunc(ctx.tracer))
	}

	// Create a router
	router, err := httprouter.NewRouter(ctx.ctx, ctx.HTTP.Prefix, ctx.HTTP.Origin, ctx.execName, version.GitTag, middleware...)
	if err != nil {
		return err
	}

	// Add handlers to the router
	httphandler.RegisterHandlers(router, "", true, manager)
	router.RegisterNotFound("/", false)
	router.RegisterOpenAPI("/openapi.json", false)

	// Create a TLS config
	var tlsconfig *tls.Config
	if cmd.TLS.CertFile != "" || cmd.TLS.KeyFile != "" {
		tlsconfig, err = httpserver.TLSConfig(cmd.TLS.ServerName, true, cmd.TLS.CertFile, cmd.TLS.KeyFile)
		if err != nil {
			return err
		}
	}

	// Create a HTTP server with timeouts
	httpopts := []httpserver.Opt{}
	if ctx.HTTP.Timeout > 0 {
		httpopts = append(httpopts, httpserver.WithReadTimeout(ctx.HTTP.Timeout))
		httpopts = append(httpopts, httpserver.WithWriteTimeout(ctx.HTTP.Timeout))
	}
	server, err := httpserver.New(ctx.HTTP.Addr, router, tlsconfig, httpopts...)
	if err != nil {
		return err
	}

	// Output the version
	if version.GitTag != "" {
		ctx.logger.Printf(ctx.ctx, "%s@%s", ctx.execName, version.GitTag)
	} else if version.GitHash != "" {
		ctx.logger.Printf(ctx.ctx, "%s@%s", ctx.execName, version.GitHash[:8])
	} else {
		ctx.logger.Printf(ctx.ctx, "%s@dev", ctx.execName)
	}

	// Run the HTTP server
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Output listening information
		ctx.logger.With("addr", ctx.HTTP.Addr).Print(ctx.ctx, "http server starting")

		// Run the server
		if err := server.Run(ctx.ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				result = errors.Join(result, fmt.Errorf("http server error: %w", err))
			}
			ctx.cancel()
		}
	}()

	// Wait for goroutine to finish
	wg.Wait()

	// Terminated message
	if result == nil {
		ctx.logger.Print(ctx.ctx, "terminated gracefully")
	} else {
		ctx.logger.Print(ctx.ctx, result)
	}

	// Return any error
	return result
}
