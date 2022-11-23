package httpserver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	// Modules

	fcgi "github.com/mutablelogic/terraform-provider-nginx/pkg/fcgi"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/terraform-provider-nginx/plugin"
)

type httpserver struct {
	Router
	label string
	srv   *http.Server
	fcgi  *fcgi.Server
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the server
func NewWithConfig(c Config) (*httpserver, error) {
	this := new(httpserver)
	this.label = c.Label

	// Set router
	if _, ok := c.Router.(http.Handler); !ok {
		return nil, ErrInternalAppError.With("invalid router")
	} else if _, ok := c.Router.(Router); !ok {
		return nil, ErrInternalAppError.With("invalid router")
	} else {
		this.Router = c.Router.(Router)
	}

	// Check addr for being (host, port). If not, then run as FCGI server
	if _, _, err := net.SplitHostPort(c.Addr); c.Addr != "" && err != nil {
		if err := this.fcgiserver(c.Addr, c.Router.(http.Handler)); err != nil {
			return nil, err
		} else {
			return this, nil
		}
	}

	// If either key or cert is non-nil then create a TLSConfig
	var tlsconfig *tls.Config
	if c.TLS != nil {
		if cert, err := tls.LoadX509KeyPair(c.TLS.Cert, c.TLS.Key); err != nil {
			return nil, err
		} else {
			tlsconfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}
	}

	// If addr is empty, then set depending on whether it's SSL or not
	if c.Addr == "" {
		if tlsconfig == nil {
			c.Addr = ":http"
		} else {
			c.Addr = ":https"
		}
	}

	// Create net server
	if err := this.netserver(c.Addr, tlsconfig, c.Timeout, c.Router.(http.Handler)); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *httpserver) String() string {
	str := "<httpserver"
	if this.label != "" {
		str += fmt.Sprintf(" label=%q", this.label)
	}
	if this.fcgi != nil {
		str += fmt.Sprintf(" fcgi=%q", this.fcgi.Addr)
	} else {
		str += fmt.Sprintf(" addr=%q", this.srv.Addr)
		if this.srv.TLSConfig != nil {
			str += " tls=true"
		}
		if this.srv.ReadHeaderTimeout != 0 {
			str += fmt.Sprintf(" read_timeout=%v", this.srv.ReadHeaderTimeout)
		}
	}
	if this.Router != nil {
		str += fmt.Sprintf(" router=%v", this.Router)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *httpserver) fcgiserver(path string, handler http.Handler) error {
	// Create server
	this.fcgi = &fcgi.Server{}
	this.fcgi.Network = "unix"
	this.fcgi.Addr = path
	this.fcgi.Handler = handler

	// Return success
	return nil
}

func (this *httpserver) netserver(addr string, config *tls.Config, timeout time.Duration, handler http.Handler) error {
	// Set up server
	this.srv = &http.Server{}
	if config != nil {
		this.srv.TLSConfig = config
	}

	// Set default timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	// Set server parameters
	this.srv.Addr = addr
	this.srv.Handler = handler
	this.srv.ReadHeaderTimeout = timeout
	this.srv.IdleTimeout = timeout

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// START AND STOP

func (this *httpserver) runInForeground() error {
	if this.fcgi != nil {
		return this.fcgi.ListenAndServe()
	} else if this.srv.TLSConfig != nil {
		return this.srv.ListenAndServeTLS("", "")
	} else {
		return this.srv.ListenAndServe()
	}
}

func (this *httpserver) stop() error {
	if this.fcgi != nil {
		return this.fcgi.Close()
	} else {
		return this.srv.Close()
	}
}
