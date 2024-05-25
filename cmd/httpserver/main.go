package main

import (
	"context"
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
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	static "github.com/mutablelogic/go-server/pkg/handler/static"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	logger "github.com/mutablelogic/go-server/pkg/middleware/logger"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

var (
	port   = flag.Int("port", 0, "Port to listen on")
	path   = flag.String("path", "", "File path to serve")
	prefix = flag.String("prefix", "/", "URL Prefix")
)

func main() {
	flag.Parse()

	// Set context
	ctx := ctx.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Create a router
	r, err := router.Config{}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create a static file handler
	filesys, err := filesys()
	if err != nil {
		log.Fatal(err)
	}
	static, err := static.Config{FS: filesys, Dir: true}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create a logger, set as middleware
	logger, err := logger.Config{Flags: []string{"default", "prefix"}}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	ctx = router.WithMiddleware(ctx, logger.(server.Middleware))

	// Add the static handler to the router
	r.(server.Router).AddHandler(router.WithPrefix(ctx, *prefix), "", static.(http.Handler))

	// Set the listen address
	listen := ":"
	if *port != 0 {
		listen = fmt.Sprintf(":%d", *port)
	}

	// Create the http server
	server, err := httpserver.Config{Listen: listen, Router: r.(http.Handler)}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Run the server until we receive a signal
	provider := provider.NewProvider(logger, static, r, server)
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
