//go:build !client

package main

import (
	"context"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	httprouter_resource "github.com/mutablelogic/go-server/pkg/httprouter/resource"
	httpserver_resource "github.com/mutablelogic/go-server/pkg/httpserver/resource"
	httpstatic "github.com/mutablelogic/go-server/pkg/httpstatic"
	httpstatic_resource "github.com/mutablelogic/go-server/pkg/httpstatic/resource"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	httphandler "github.com/mutablelogic/go-server/pkg/provider/httphandler"
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

func (r *RunServer) Run(ctx *cmd.Global) error {
	// Create the provider manager for kaiak's resource API
	manager, err := provider.New(ctx.Name(), "kaiak provider API", ctx.Version())
	if err != nil {
		return err
	}
	defer manager.Close(context.Background())

	// Register kaiak-specific resource types for dynamic creation via the API
	if err := manager.RegisterResource(httprouter_resource.Resource{}); err != nil {
		return err
	}
	if err := manager.RegisterResource(httpserver_resource.Resource{}); err != nil {
		return err
	}
	if err := manager.RegisterResource(httpstatic_resource.Resource{}); err != nil {
		return err
	}

	// Register kaiak-specific routes
	r.RunServer.Register(func(router *httprouter.Router, g server.Cmd) error {
		router.Spec().AddTag("Resources", "Resource lifecycle operations")
		if err := router.RegisterFunc("resource", httphandler.ResourceListHandler(manager), true, httphandler.ResourceListSpec()); err != nil {
			return err
		}
		return router.RegisterFunc("resource/{id}", httphandler.ResourceInstanceHandler(manager), true, httphandler.ResourceInstanceSpec())
	})

	if r.Static.Dir != "" {
		r.RunServer.Register(r.registerStaticFiles)
	}

	return r.RunServer.Run(ctx)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *RunServer) registerStaticFiles(router *httprouter.Router, g server.Cmd) error {
	fs := os.DirFS(r.Static.Dir)
	h, err := httpstatic.New(r.Static.Path, fs, "Static files", "")
	if err != nil {
		return err
	}
	return router.RegisterFS(h.HandlerPath(), h.HandlerFS(), true, h.Spec())
}
