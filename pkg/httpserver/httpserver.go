package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"
	fcgi "github.com/mutablelogic/go-server/pkg/httpserver/fcgi"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// http server configuration
type Config struct {
	Label  hcl.Label // Block Label
	Listen string    `hcl:"listen" description:"Network address and port to listen on, or path to file socket"`
	TLS    struct {
		Key  string `hcl:"key" description:"Path to private key (if provided, tls.cert is also required)"`
		Cert string `hcl:"cert" description:"Path to certificate (if provided, tls.key is also required)"`
	} `hcl:"tls"`
	Timeout time.Duration `hcl:"timeout" description:"Read request timeout"`
	Owner   string        `hcl:"owner" description:"User ID of the file socket (if listen is a file socket)"`
	Group   string        `hcl:"group" description:"Group ID of the file socket (if listen is a file socket)"`
	Router  hcl.Resource  `hcl:"router,required" description:"router for requests (http.Handler)"`
}

// http server instance
type httpserver struct {
	name  string
	label hcl.Label
	fcgi  *fcgi.Server
	http  *http.Server
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

func (Config) Name() string {
	return defaultName
}

func (Config) Description() string {
	return "serves requests and provides responses over HTTP, HTTPS and FCGI"
}

func (c Config) New(context.Context) (hcl.Resource, error) {
	self := new(httpserver)
	self.name = c.Name()
	self.label = c.Label

	router, ok := c.Router.(http.Handler)
	if router == nil {
		return nil, ErrBadParameter.Withf("Missing http.Handler")
	} else if !ok {
		return nil, ErrBadParameter.Withf("Not a http.Handler: %q", c.Router)
	}

	if c.isHostPort() {
		// Read the TLS configuration
		tls, err := c.tls()
		if err != nil {
			return nil, err
		}

		// Create net server
		if http, err := netserver(c.listen(), tls, c.timeout(), router); err != nil {
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
		} else if fcgi, err := fcgiserver(abs, owner, group, c.mode(), router); err != nil {
			return nil, err
		} else {
			self.fcgi = fcgi
		}
	}

	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Config) String() string {
	str := "<block "
	str += fmt.Sprintf("name=%q", c.Name())
	if !c.Label.IsZero() {
		str += fmt.Sprintf(" label=%q", c.Label)
	}
	if c.Listen != "" {
		str += fmt.Sprintf(" listen=%q", c.Listen)
	}
	if c.TLS.Cert != "" || c.TLS.Key != "" {
		str += fmt.Sprintf(" tls=%q", c.TLS)
	}
	if c.Timeout != 0 {
		str += fmt.Sprintf(" timeout=%v", c.Timeout)
	}
	if c.Owner != "" {
		str += fmt.Sprintf(" owner=%q", c.Owner)
	}
	if c.Group != "" {
		str += fmt.Sprintf(" group=%q", c.Group)
	}
	if c.Router != nil {
		str += fmt.Sprintf(" router=%v", c.Router)
	}
	return str + ">"
}

func (self *httpserver) String() string {
	str := "<" + self.name
	if len(self.label) > 0 {
		str += fmt.Sprintf(" label=%q", self.label)
	}
	if self.fcgi != nil {
		str += fmt.Sprintf(" fcgi=%q", self.fcgi.Addr)
	} else {
		str += fmt.Sprintf(" addr=%q", self.http.Addr)
		if self.http.TLSConfig != nil {
			str += " tls=true"
		}
		if self.http.ReadHeaderTimeout != 0 {
			str += fmt.Sprintf(" read_timeout=%v", self.http.ReadHeaderTimeout)
		}
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

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

// Return the default listen address
func (c Config) listen() string {
	if c.Listen == "" {
		if c.TLS.Cert != "" || c.TLS.Key != "" {
			return defaultListenTLS
		} else {
			return defaultListen
		}
	} else {
		return c.Listen
	}
}

// TODO: return true if:
//
//	:<number|name> <ip4>:<number> <ip6>:<number> <iface>:<number|name>
func (c Config) isHostPort() bool {
	// Empty string not acceptable
	if c.Listen == "" {
		return false
	}
	// Check for host:port and if no error, return true
	if _, _, err := net.SplitHostPort(c.Listen); err == nil {
		return true
	} else {
		return false
	}
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
