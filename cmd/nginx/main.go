package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	server "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	logger "github.com/mutablelogic/go-server/pkg/middleware/logger"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

var (
	binary = flag.String("path", "nginx", "Path to nginx binary")
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

	// Nginx handler
	n, err := nginx.Config{BinaryPath: *binary}.New()
	if err != nil {
		log.Fatal(err)
	}
	socket := filepath.Join(n.(nginx.Nginx).Config(), "go-server.sock")

	// Print the location of the nginx configuration files
	fmt.Println(socket)

	// Router
	router, err := router.Config{
		Services: router.ServiceConfig{
			"nginx": { // /api/nginx/...
				Service: n.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
				},
			},
		},
	}.New()
	if err != nil {
		log.Fatal(err)
	}

	// HTTP Server
	httpserver, err := httpserver.Config{Listen: socket, Router: router.(http.Handler)}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Run until we receive an interrupt
	provider := provider.NewProvider(logger, n, router, httpserver)
	provider.Print(ctx, "Press CTRL+C to exit")
	if err := provider.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
