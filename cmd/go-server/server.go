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
	// Create the manager
	manager, err := provider.New(ctx.execName, "A generic server for running providers", version.GitTag)
	if err != nil {
		return err
	}

	// Compute a version string (ldflags may not be set in dev builds)
	versionTag := version.GitTag
	if versionTag == "" && version.GitHash != "" {
		versionTag = version.GitHash[:8]
	} else if versionTag == "" {
		versionTag = "dev"
	}

	// Build ordered middleware and handler name lists for the router state
	var middlewareNames []string

	// Create a read-only logger instance
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

	// Create a read-only otel middleware instance (passive — no dependencies)
	if ctx.tracer != nil {
		otelInst, err := manager.RegisterReadonlyInstance(ctx.ctx, otel_resource.NewResource(otel.HTTPHandlerFunc(ctx.tracer)), "main", schema.State{})
		if err != nil {
			return err
		}
		middlewareNames = append(middlewareNames, otelInst.Name())
	}

	// Create read-only handler instances (passive — no dependencies)
	handlerResources := []handler_resource.Resource{
		handler_resource.NewResource(
			"provider-api", "resource",
			httphandler.ResourceListHandler(manager),
			httphandler.ResourceListSpec(),
		),
		handler_resource.NewResource(
			"provider-api-item", "resource/{id}",
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
	routerInst, err := manager.RegisterReadonlyInstance(ctx.ctx, httprouter_resource.Resource{}, "main", schema.State{
		"prefix":     ctx.HTTP.Prefix,
		"origin":     ctx.HTTP.Origin,
		"title":      ctx.execName,
		"version":    versionTag,
		"openapi":    cmd.OpenAPI,
		"middleware": middlewareNames,
		"handlers":   handlerNames,
	})
	if err != nil {
		return err
	}

	// Create a read-only httpserver instance
	serverState := schema.State{
		"listen": ctx.HTTP.Addr,
		"router": routerInst.Name(),
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
		serverState["read-timeout"] = ctx.HTTP.Timeout.String()
		serverState["write-timeout"] = ctx.HTTP.Timeout.String()
	}

	// Create the read-only httpserver instance, which starts the server and serves requests
	httpserver, err := manager.RegisterReadonlyInstance(ctx.ctx, httpserver_resource.Resource{}, "main", serverState)
	if err != nil {
		return err
	}

	// Output the version
	ctx.logger.Printf(ctx.ctx, "%s@%s", ctx.execName, versionTag)
	if state, err := httpserver.Read(ctx.ctx); err == nil {
		ctx.logger.Printf(ctx.ctx, "Listening on %v", state["endpoint"])
	}

	// Wait for context cancellation (e.g. signal), then close all resources
	<-ctx.ctx.Done()

	// Close the manager, destroying all resources in dependency order
	// (httpserver stops first, then httprouter is torn down)
	if err := manager.Close(ctx.ctx); err != nil {
		return err
	}

	// Terminated message
	ctx.logger.Print(ctx.ctx, "terminated gracefully")

	// Return any error
	return nil
}
