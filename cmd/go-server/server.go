//go:build !client

package main

import (
	"fmt"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprouter_resource "github.com/mutablelogic/go-server/pkg/httprouter/resource"
	httpserver_resource "github.com/mutablelogic/go-server/pkg/httpserver/resource"
	httpstatic_resource "github.com/mutablelogic/go-server/pkg/httpstatic/resource"
	log_resource "github.com/mutablelogic/go-server/pkg/logger/resource"
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
	OpenAPI bool `name:"openapi" help:"Serve OpenAPI spec at {prefix}/openapi.json" default:"true" negatable:""`

	// Static file serving options
	Static struct {
		Path string `name:"path" help:"URL path for static files (relative to prefix)" default:""`
		Dir  string `name:"dir" help:"Directory on disk to serve as static files" default:""`
	} `embed:"" prefix:"static."`

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
	return cmd.WithManager(ctx, func(manager *provider.Manager, v string) error {
		// Start the HTTP server and wait for shutdown
		return cmd.Serve(ctx, manager, version.Version())
	})
}

// withManager creates the resource manager, registers all resource instances
// (logger, otel, handlers, router) in dependency order, invokes fn, then
// closes the manager regardless of whether fn returned an error.
func (cmd *RunServer) WithManager(ctx *Globals, fn func(*provider.Manager, string) error) error {
	manager, err := provider.New(ctx.execName, "A generic server for running providers", version.Version())
	if err != nil {
		return err
	}
	defer manager.Close(ctx.ctx)

	// Create a read-only logger instance
	var middlewareNames []string
	format := "text"
	if isTerminal(os.Stderr) {
		format = "term"
	}
	logInst, err := manager.RegisterReadonlyInstance(ctx.ctx, log_resource.Resource{}, "main", schema.State{
		"debug":  ctx.Debug,
		"format": format,
	})
	if err != nil {
		return err
	}
	middlewareNames = append(middlewareNames, logInst.Name())

	// Replace the early bootstrap logger with the resource instance
	if l, ok := logInst.(server.Logger); ok {
		ctx.logger = l
	}

	// Create a read-only otel middleware instance
	if ctx.tracer != nil {
		otelInst, err := manager.RegisterReadonlyInstance(ctx.ctx, otel_resource.NewResource(otel.HTTPHandlerFunc(ctx.tracer)), "main", schema.State{})
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
		inst, err := manager.RegisterReadonlyInstance(ctx.ctx, r, "main", schema.State{})
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
	if cmd.Static.Dir != "" {
		staticInst, err := manager.RegisterReadonlyInstance(ctx.ctx, httpstatic_resource.Resource{}, "main", schema.State{
			"path": cmd.Static.Path,
			"dir":  cmd.Static.Dir,
		})
		if err != nil {
			return err
		}
		handlerNames = append(handlerNames, staticInst.Name())
	}

	// Create a read-only httprouter instance — references middleware and handlers
	if _, err := manager.RegisterReadonlyInstance(ctx.ctx, httprouter_resource.Resource{}, "main", schema.State{
		"prefix":     ctx.HTTP.Prefix,
		"origin":     ctx.HTTP.Origin,
		"title":      ctx.execName,
		"version":    version.Version(),
		"openapi":    cmd.OpenAPI,
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
func (cmd *RunServer) Serve(ctx *Globals, manager *provider.Manager, versionTag string) error {
	// Build the httpserver state
	serverState := schema.State{
		"listen": ctx.HTTP.Addr,
		"router": "httprouter.main",
	}
	if cmd.TLS.CertFile != "" || cmd.TLS.KeyFile != "" {
		certPEM, err := os.ReadFile(cmd.TLS.CertFile)
		if err != nil {
			return fmt.Errorf("tls.cert: %w", err)
		}
		keyPEM, err := os.ReadFile(cmd.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("tls.key: %w", err)
		}
		serverState["tls.cert"] = certPEM
		serverState["tls.key"] = keyPEM
		serverState["tls.name"] = cmd.TLS.ServerName
	}
	if ctx.HTTP.Timeout > 0 {
		serverState["read_timeout"] = ctx.HTTP.Timeout.String()
		serverState["write_timeout"] = ctx.HTTP.Timeout.String()
	}

	// Create the read-only httpserver instance, which starts the server
	httpserver, err := manager.RegisterReadonlyInstance(ctx.ctx, httpserver_resource.Resource{}, "main", serverState)
	if err != nil {
		return err
	}

	// Output the version and endpoint
	ctx.logger.Printf(ctx.ctx, "%s@%s", ctx.execName, versionTag)
	if state, err := httpserver.Read(ctx.ctx); err == nil {
		ctx.logger.Printf(ctx.ctx, "Listening on %v", state["endpoint"])
	}

	// Wait for context cancellation (e.g. signal)
	<-ctx.ctx.Done()

	// Log before closing — manager.Close destroys the logger
	ctx.logger.Print(ctx.ctx, "terminated gracefully")

	// Return success
	return nil
}
