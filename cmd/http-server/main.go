package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"syscall"

	// Packages
	server "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	logger "github.com/mutablelogic/go-server/pkg/handler/logger"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	static "github.com/mutablelogic/go-server/pkg/handler/static"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

var (
	port   = flag.Int("port", 0, "Port to listen on")
	path   = flag.String("path", "", "File path to serve")
	prefix = flag.String("prefix", "/", "URL Prefix")
)

func main() {
	flag.Parse()

	// Create context which cancels on interrupt
	ctx := ctx.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Logger
	logger, err := logger.Config{Flags: []string{"default", "prefix"}}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Static file handler
	filesys, err := filesys()
	if err != nil {
		log.Fatal(err)
	}
	static, err := static.Config{FS: filesys, DirListing: true}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Router
	router, err := router.Config{
		Services: router.ServiceConfig{
			*prefix: {
				Service: static.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
				},
			},
		},
	}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Set the listen address
	listen := ":"
	if *port != 0 {
		listen = fmt.Sprintf(":%d", *port)
	}

	// HTTP Server
	httpserver, err := httpserver.Config{Listen: listen, Router: router.(http.Handler)}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Run until we receive an interrupt
	provider := provider.NewProvider(logger, static, router, httpserver)
	provider.Print(ctx, "Press CTRL+C to exit")
	if err := provider.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func filesys() (fs.FS, error) {
	if *path != "" {
		return os.DirFS(*path), nil
	} else if wd, err := os.Getwd(); err != nil {
		return nil, err
	} else {
		return os.DirFS(wd), nil
	}
}
