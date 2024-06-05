package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth"
	logger "github.com/mutablelogic/go-server/pkg/handler/logger"
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	tokenjar "github.com/mutablelogic/go-server/pkg/handler/tokenjar"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

var (
	binary = flag.String("path", "nginx", "Path to nginx binary")
	group  = flag.String("group", "", "Group to run unix socket as")
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

	// Token Jar
	jar, err := tokenjar.Config{
		DataPath:      n.(nginx.Nginx).Config(),
		WriteInterval: 30 * time.Second,
	}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Auth handler
	auth, err := auth.Config{
		TokenJar:   jar.(auth.TokenJar),
		TokenBytes: 8,
		Bearer:     true, // Use bearer token in requests for authorization
	}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Location of the FCGI unix socket
	socket := filepath.Join(n.(nginx.Nginx).Config(), "run/go-server.sock")

	// Router
	router, err := router.Config{
		Services: router.ServiceConfig{
			"nginx": { // /api/nginx/...
				Service: n.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
			"auth": { // /api/auth/...
				Service: auth.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
		},
	}.New()
	if err != nil {
		log.Fatal(err)
	}

	// HTTP Server
	httpserver, err := httpserver.Config{
		Listen: socket,
		Group:  *group,
		Router: router.(http.Handler),
	}.New()
	if err != nil {
		log.Fatal(err)
	}

	// Run until we receive an interrupt
	provider := provider.NewProvider(logger, n, jar, auth, router, httpserver)
	provider.Print(ctx, "Press CTRL+C to exit")
	if err := provider.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
