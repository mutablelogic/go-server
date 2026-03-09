//go:build !client

package main

import (
	"context"
	"fmt"
	"os"

	// Packages
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	httprouter_resource "github.com/mutablelogic/go-server/pkg/httprouter/resource"
	httpserver_resource "github.com/mutablelogic/go-server/pkg/httpserver/resource"
	httpstatic_resource "github.com/mutablelogic/go-server/pkg/httpstatic/resource"
	otel "github.com/mutablelogic/go-server/pkg/otel"
	otel_resource "github.com/mutablelogic/go-server/pkg/otel/resource"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	httphandler "github.com/mutablelogic/go-server/pkg/provider/httphandler"
	handler_resource "github.com/mutablelogic/go-server/pkg/provider/httphandler/resource"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run server." group:"SERVER"`
}

type RunServer struct {
	cmd.RunServer

	// Static file serving options
	Static struct {
		Path string `name:"path" help:"URL path for static files (relative to prefix)" default:""`
		Dir  string `name:"dir" help:"Directory on disk to serve as static files" default:""`
	} `embed:"" prefix:"static."`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (s *RunServer) Run(ctx *cmd.Global) error {
	return s.WithManager(ctx, func(manager *provider.Manager, v string) error {
		// Start the HTTP server and wait for shutdown
		return s.Serve(ctx, manager, version.Version())
	})
}

// withManager creates the resource manager, registers all resource instances
// (logger, otel, handlers, router) in dependency order, invokes fn, then
// closes the manager regardless of whether fn returned an error.
func (s *RunServer) WithManager(ctx *cmd.Global, fn func(*provider.Manager, string) error) error {
	manager, err := provider.New(ctx.Name(), "A generic server for running providers", version.Version())
	if err != nil {
		return err
	}
	defer manager.Close(ctx.Context())

	// logger middleware comes from ctx.Logger() — no separate resource needed
	var middlewareNames []string

	// Create a read-only otel middleware instance
	if ctx.Tracer() != nil {
		otelInst, err := manager.RegisterReadonlyInstance(ctx.Context(), otel_resource.NewResource(otel.HTTPHandlerFunc(ctx.Tracer())), "main", schema.State{})
		if err != nil {
			return err
		}
		middlewareNames = append(middlewareNames, otelInst.Name())
	}

	// Create read-only handler instances
	handlerResources := []handler_resource.Resource{
		handler_resource.NewResource(
			"provider_api_resources", "resource",
			httphandler.ResourceListHandler(manager),
			httphandler.ResourceListSpec(),
		),
		handler_resource.NewResource(
			"provider_api_instances", "resource/{id}",
			httphandler.ResourceInstanceHandler(manager),
			httphandler.ResourceInstanceSpec(),
		),
	}
	handlerNames := make([]string, 0, len(handlerResources))
	for _, r := range handlerResources {
		inst, err := manager.RegisterReadonlyInstance(ctx.Context(), r, "main", schema.State{})
		if err != nil {
			return err
		}
		handlerNames = append(handlerNames, inst.Name())
	}

	// Register the httpstatic resource type so it can be created dynamically
	// via the API, and optionally create a read-only instance if a static
	// directory is configured on the command line.
	if err := manager.RegisterResource(httpstatic_resource.Resource{}); err != nil {
		return err
	}
	if s.Static.Dir != "" {
		staticInst, err := manager.RegisterReadonlyInstance(ctx.Context(), httpstatic_resource.Resource{}, "main", schema.State{
			"path": s.Static.Path,
			"dir":  s.Static.Dir,
		})
		if err != nil {
			return err
		}
		handlerNames = append(handlerNames, staticInst.Name())
	}

	// Create a read-only httprouter instance — references middleware and handlers
	if _, err := manager.RegisterReadonlyInstance(ctx.Context(), httprouter_resource.Resource{}, "main", schema.State{
		"prefix":     ctx.HTTP.Prefix,
		"origin":     ctx.HTTP.Origin,
		"title":      ctx.Name(),
		"version":    version.Version(),
		"openapi":    s.OpenAPI,
		"middleware": middlewareNames,
		"handlers":   handlerNames,
	}); err != nil {
		return err
	}

	return fn(manager, version.Version())
}

// Serve creates the httpserver instance, logs the startup banner, and
// blocks until context cancellation (e.g. SIGINT). The caller is
// responsible for closing the manager afterwards.
func (s *RunServer) Serve(ctx *cmd.Global, manager *provider.Manager, versionTag string) error {
	// Build the httpserver state
	serverState := schema.State{
		"listen": ctx.HTTP.Addr,
		"router": "httprouter.main",
	}
	if s.TLS.CertFile != "" || s.TLS.KeyFile != "" {
		certPEM, err := os.ReadFile(s.TLS.CertFile)
		if err != nil {
			return fmt.Errorf("tls.cert: %w", err)
		}
		keyPEM, err := os.ReadFile(s.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("tls.key: %w", err)
		}
		serverState["tls.cert"] = certPEM
		serverState["tls.key"] = keyPEM
		serverState["tls.name"] = s.TLS.ServerName
	}
	if ctx.HTTP.Timeout > 0 {
		serverState["read_timeout"] = ctx.HTTP.Timeout.String()
		serverState["write_timeout"] = ctx.HTTP.Timeout.String()
	}

	// Create the read-only httpserver instance, which starts the server
	httpserver, err := manager.RegisterReadonlyInstance(ctx.Context(), httpserver_resource.Resource{}, "main", serverState)
	if err != nil {
		return err
	}

	// Output the version and endpoint
	ctx.Logger().InfoContext(ctx.Context(), ctx.Name(), "version", versionTag)
	if state, err := httpserver.Read(ctx.Context()); err == nil {
		ctx.Logger().InfoContext(ctx.Context(), "listening", "endpoint", state["endpoint"])
	}

	// Wait for context cancellation (e.g. signal)
	<-ctx.Context().Done()

	// Log before closing — manager.Close destroys the logger
	ctx.Logger().InfoContext(context.Background(), "terminated gracefully")

	// Return success
	return nil
}
