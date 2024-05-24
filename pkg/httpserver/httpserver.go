package httpserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	// Packages
	fcgi "github.com/mutablelogic/go-server/pkg/httpserver/fcgi"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// http server configuration
type Config struct {
	Listen string `hcl:"listen" description:"Network address and port to listen on, or path to file socket"`
	TLS    struct {
		Key  string `hcl:"key" description:"Path to private key (if provided, tls.cert is also required)"`
		Cert string `hcl:"cert" description:"Path to certificate (if provided, tls.key is also required)"`
	} `hcl:"tls"`
	Timeout time.Duration `hcl:"timeout" description:"Read request timeout"`
	Owner   string        `hcl:"owner" description:"User ID of the file socket (if listen is a file socket)"`
	Group   string        `hcl:"group" description:"Group ID of the file socket (if listen is a file socket)"`
	Router  http.Handler  `hcl:"router" description:"HTTP router for requests"`
}

// http server instance
type httpserver struct {
	fcgi *fcgi.Server
	http *http.Server
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName      = "httpserver"
	defaultTimeout   = 10 * time.Second
	defaultListen    = ":http"
	defaultListenTLS = ":https"
	defaultMode      = os.FileMode(0600)
	groupMode        = os.FileMode(0060)
	allMode          = os.FileMode(0777)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Name returns the name of the service
func (Config) Name() string {
	return defaultName
}

// Description returns the description of the service
func (Config) Description() string {
	return "serves requests and provides responses over HTTP, HTTPS and FCGI"
}

// Create a new http server from the configuration
func (c Config) New(context.Context) (*httpserver, error) {
	self := new(httpserver)

	// Create a default router if not provided
	if c.Router == nil {
		c.Router = http.DefaultServeMux
	}

	// Set defaults
	if c.Listen == "" {
		if c.TLS.Cert != "" || c.TLS.Key != "" {
			c.Listen = defaultListenTLS
		} else {
			c.Listen = defaultListen
		}
	}

	// Choose HTTP or FCGI server
	if strings.ContainsRune(c.Listen, ':') {
		host, port, err := c.isHostPort()
		if err != nil {
			return nil, err
		}

		// Read the TLS configuration
		tls, err := c.tls()
		if err != nil {
			return nil, err
		}

		// Create net server
		addr := fmt.Sprintf("%s:%d", host, port)
		if http, err := netserver(addr, tls, c.timeout(), c.Router); err != nil {
			return nil, err
		} else {
			self.http = http
		}
	} else if abs, err := filepath.Abs(c.Listen); err != nil {
		return nil, err
	} else {
		// file socket parameters
		if owner, err := c.owner(); err != nil {
			return nil, err
		} else if group, err := c.group(); err != nil {
			return nil, err
		} else if fcgi, err := fcgiserver(abs, owner, group, c.mode(), c.Router); err != nil {
			return nil, err
		} else {
			self.fcgi = fcgi
		}
	}

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return the configuration as a string
func (c Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func (h httpserver) String() string {
	var server struct {
		Type string `json:"type"`
		Addr string `json:"addr"`
	}
	if h.fcgi != nil {
		server.Type = "fcgi"
		server.Addr = h.fcgi.Addr
	}
	if h.http != nil {
		server.Type = "http"
		server.Addr = h.http.Addr
	}
	data, _ := json.MarshalIndent(server, "", "  ")
	return string(data)
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run the http server until the context is cancelled
func (self *httpserver) Run(ctx context.Context) error {
	var result error
	var wg sync.WaitGroup

	// Create cancelable context
	child, cancel := context.WithCancel(ctx)
	defer cancel()

	// Stop server on context cancel
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-child.Done()
		if err := self.stop(); err != nil {
			result = errors.Join(result, err)
		}
	}()

	// Run server in foreground, cancel when done
	if err := self.runInForeground(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		result = errors.Join(result, err)
		cancel()
	}

	// Wait for gorutines to finish
	wg.Wait()

	// Return any errors
	return result
}

// Return the router for the server
func (self *httpserver) Router() http.Handler {
	if self.fcgi != nil {
		return self.fcgi.Handler
	} else {
		return self.http.Handler
	}
}

// Return the type of server
func (self *httpserver) Type() string {
	if self.fcgi != nil {
		return "fcgi"
	}
	return "http"
}

// Return the listening address or path
func (self *httpserver) Addr() string {
	if self.fcgi != nil {
		return self.fcgi.Addr
	}
	return self.http.Addr
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return the timeout value from the configuration
func (c Config) timeout() time.Duration {
	if c.Timeout != 0 {
		return c.Timeout
	} else {
		return defaultTimeout
	}
}

// Return the TLS configuration
func (c Config) tls() (*tls.Config, error) {
	if c.TLS.Cert == "" || c.TLS.Key == "" {
		return nil, nil
	}
	if cert, err := tls.LoadX509KeyPair(c.TLS.Cert, c.TLS.Cert); err != nil {
		return nil, fmt.Errorf("LoadX509KeyPair: %w", err)
	} else {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}
}

// Return the uid of the socket owner
func (c Config) owner() (int, error) {
	if c.Owner == "" {
		return -1, nil
	}
	if user, err := user.Lookup(c.Owner); err != nil {
		return -1, err
	} else if uid, err := strconv.ParseUint(user.Uid, 0, 32); err != nil {
		return -1, err
	} else {
		return int(uid), nil
	}
}

// Return the gid of the socket group
func (c Config) group() (int, error) {
	if c.Group == "" {
		return -1, nil
	}
	if group, err := user.LookupGroup(c.Group); err != nil {
		return -1, err
	} else if gid, err := strconv.ParseUint(group.Gid, 0, 32); err != nil {
		return -1, err
	} else {
		return int(gid), nil
	}
}

// Return the socket mode
func (c Config) mode() os.FileMode {
	mode := defaultMode
	if gid, err := c.group(); gid != -1 && err == nil {
		mode |= groupMode
	}
	return mode & allMode
}

// Returns the host and port number from the configuration
// Random port number is chosen if port is not provided or is zero
func (c Config) isHostPort() (string, int, error) {
	// Empty string not acceptable
	if c.Listen == "" {
		return "", 0, ErrBadParameter.With("empty listen address")
	}
	// Check for host:port and if no error, return true
	if host, port, err := net.SplitHostPort(c.Listen); err != nil {
		return "", 0, err
	} else if port == "" || port == "0" {
		if port, err := getFreePort(); err != nil {
			return "", 0, err
		} else {
			return host, port, nil
		}
	} else if port_, err := strconv.ParseUint(port, 10, 16); err != nil {
		if port_, err := net.LookupPort("tcp", port); err != nil {
			return "", 0, err
		} else {
			return host, port_, nil
		}
	} else {
		return host, int(port_), nil
	}
}

// getFreePort returns a random free port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Run the server and block until stopped
func (self *httpserver) runInForeground() error {
	if self.fcgi != nil {
		return self.fcgi.ListenAndServe()
	} else if self.http.TLSConfig != nil {
		return self.http.ListenAndServeTLS("", "")
	} else {
		return self.http.ListenAndServe()
	}
}

// Stop the server
func (self *httpserver) stop() error {
	if self.fcgi != nil {
		return self.fcgi.Close()
	} else {
		return self.http.Close()
	}
}

// Create a fastcgi file socket server
func fcgiserver(path string, uid, gid int, mode os.FileMode, handler http.Handler) (*fcgi.Server, error) {
	fcgi := new(fcgi.Server)
	fcgi.Network = "unix"
	fcgi.Addr = path
	fcgi.Handler = handler
	fcgi.Owner = uid
	fcgi.Group = gid
	fcgi.Mode = mode

	// Return success
	return fcgi, nil
}

// Create a network socket server
func netserver(addr string, config *tls.Config, timeout time.Duration, handler http.Handler) (*http.Server, error) {
	srv := new(http.Server)
	srv.Addr = addr
	srv.Handler = handler
	srv.ReadHeaderTimeout = timeout
	srv.IdleTimeout = timeout
	if config != nil {
		srv.TLSConfig = config
	}

	// Return success
	return srv, nil
}
