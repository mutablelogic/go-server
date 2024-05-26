package main

import (
	"flag"
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

/* command to test the nginx package */
/* will run the nginx server and provide an nginx api for reloading,
   testing, etc through FastCGI. The config and run paths are a bit
   screwed up and will need to be fixed.
*/
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

	// Location of the FCGI unix socket
	socket := filepath.Join(n.(nginx.Nginx).Config(), "go-server.sock")

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
