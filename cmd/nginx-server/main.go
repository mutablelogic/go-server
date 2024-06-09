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
	routerFrontend "github.com/mutablelogic/go-server/npm/router-frontend"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth"
	certmanager "github.com/mutablelogic/go-server/pkg/handler/certmanager"
	certstore "github.com/mutablelogic/go-server/pkg/handler/certmanager/certstore"
	ldap "github.com/mutablelogic/go-server/pkg/handler/ldap"
	logger "github.com/mutablelogic/go-server/pkg/handler/logger"
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	tokenjar "github.com/mutablelogic/go-server/pkg/handler/tokenjar"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

var (
	binary        = flag.String("nginx", "nginx", "Path to nginx binary")
	group         = flag.String("group", "", "Group to run unix socket as")
	data          = flag.String("data", "", "Path to data (emphermeral) directory")
	conf          = flag.String("conf", "", "Path to conf (persistent) directory")
	ldap_password = flag.String("ldap-password", "", "LDAP admin password")
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

	// Set of tasks
	var tasks []server.Task

	// Logger
	logger, err := logger.Config{Flags: []string{"default", "prefix"}}.New()
	if err != nil {
		log.Fatal("logger: ", err)
	} else {
		tasks = append(tasks, logger)
	}

	// Nginx handler
	n, err := nginx.Config{
		BinaryPath: *binary,
		DataPath:   *data,
		ConfigPath: *conf,
	}.New()
	if err != nil {
		log.Fatal("nginx: ", err)
	} else {
		tasks = append(tasks, n)
	}

	// Token Jar
	jar, err := tokenjar.Config{
		DataPath:      n.(nginx.Nginx).ConfigPath(),
		WriteInterval: 30 * time.Second,
	}.New()
	if err != nil {
		log.Fatal("tokenjar: ", err)
	} else {
		tasks = append(tasks, jar)
	}

	// Auth handler
	auth, err := auth.Config{
		TokenJar:   jar.(auth.TokenJar),
		TokenBytes: 8,
		Bearer:     true, // Use bearer token in requests for authorization
	}.New()
	if err != nil {
		log.Fatal("auth: ", err)
	} else {
		tasks = append(tasks, auth)
	}

	// Cert storage
	certstore, err := certstore.Config{
		DataPath: filepath.Join(n.(nginx.Nginx).ConfigPath(), "cert"),
		Group:    *group,
	}.New()
	if err != nil {
		log.Fatal("certstore: ", err)
	} else {
		tasks = append(tasks, certstore)
	}

	// Cert manager
	certmanager, err := certmanager.Config{
		CertStorage: certstore.(certmanager.CertStorage),
		X509Name: certmanager.X509Name{
			OrganizationalUnit: "mutablelogic.com",
			Organization:       "mutablelogic",
			StreetAddress:      "N/A",
			Locality:           "Berlin",
			Province:           "Berlin",
			PostalCode:         "10967",
			Country:            "DE",
		},
	}.New()
	if err != nil {
		log.Fatal("certmanager: ", err)
	} else {
		tasks = append(tasks, certmanager)
	}

	// Router frontend
	routerFrontend, err := routerFrontend.Config{}.New()
	if err != nil {
		log.Fatal("routerFrontend: ", err)
	} else {
		tasks = append(tasks, routerFrontend)
	}

	// Router
	// TODO: Promote middleware to the root of the configuration to reduce
	// duplication
	r, err := router.Config{
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
			"cert": { // /api/cert/...
				Service: certmanager.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
		},
	}.New()
	if err != nil {
		log.Fatal("router: ", err)
	} else {
		tasks = append(tasks, r)
	}

	// Add router and frontend
	// The API is served from http[s]://[any]/api/router and the frontend is served from http[s]://static/router
	// see the nginx default configuration to understand how the routing occurs in the proxy
	r.(router.Router).AddServiceEndpoints("router", r.(server.ServiceEndpoints), logger.(server.Middleware), auth.(server.Middleware))
	r.(router.Router).AddServiceEndpoints("static/router", routerFrontend.(server.ServiceEndpoints), logger.(server.Middleware))

	// LDAP gets enabled if a password is set
	if *ldap_password != "" {
		ldap, err := ldap.Config{
			URL:      "ldap://admin@cm1.local/",
			DN:       "dc=mutablelogic,dc=com",
			Password: *ldap_password,
		}.New()
		if err != nil {
			log.Fatal("ldap: ", err)
		} else {
			r.(router.Router).AddServiceEndpoints("ldap", ldap.(server.ServiceEndpoints), logger.(server.Middleware), auth.(server.Middleware))
			tasks = append(tasks, ldap)
		}
	}

	// Location of the FCGI unix socket - this should be the same
	// as that listed in the nginx configuration
	socket := filepath.Join(n.(nginx.Nginx).DataPath(), "nginx/go-server.sock")

	// HTTP Server
	httpserver, err := httpserver.Config{
		Listen: socket,
		Group:  *group,
		Router: r.(http.Handler),
	}.New()
	if err != nil {
		log.Fatal("httpserver: ", err)
	} else {
		tasks = append(tasks, httpserver)
	}

	// Run until we receive an interrupt
	provider := provider.NewProvider(tasks...)
	provider.Print(ctx, "Press CTRL+C to exit")
	if err := provider.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
