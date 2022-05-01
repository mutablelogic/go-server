package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-server"
	fcgi "github.com/mutablelogic/go-server/pkg/fcgi"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	prv "github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Label   string `hcl:"label,label"` // Label for HTTP server
	private string
	Addr    string        `hcl:"addr,optional"`    // Address or path for binding HTTP server
	TLS     *TLS          `hcl:"tls,block"`        // TLS parameters
	Timeout time.Duration `hcl:"timeout,optional"` // Read timeout on HTTP requests
}

type TLS struct {
	Key  string `hcl:"key"`  // Path to TLS Private Key
	Cert string `hcl:"cert"` // Path to TLS Certificate
}

type server struct {
	Config
	Router
	srv  *http.Server
	fcgi *fcgi.Server
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultTimeout = 10 * time.Second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(server)
	this.Router = httprouter.New()

	// Get configuration
	if err := provider.GetConfig(ctx, &this.Config); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Override addr if set in context
	if addr := prv.ContextAddr(ctx); addr != "" {
		this.Config.Addr = addr
	}

	// Check addr for being (host,port). If not, then run as FCGI server
	if _, _, err := net.SplitHostPort(this.Config.Addr); this.Config.Addr != "" && err != nil {
		if err := this.fcgiserver(this.Config.Addr); err != nil {
			provider.Print(ctx, err)
			return nil
		}
	}

	// If either key or cert is non-nil then create a TLSConfig
	var tlsconfig *tls.Config
	if this.Config.TLS != nil {
		if cert, err := tls.LoadX509KeyPair(this.Config.TLS.Cert, this.Config.TLS.Key); err != nil {
			provider.Print(ctx, err)
			return nil
		} else {
			tlsconfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}
	}

	// If addr is empty, then set depending on whether it's SSL or not
	if this.Config.Addr == "" {
		if tlsconfig == nil {
			this.Config.Addr = ":http"
		} else {
			this.Config.Addr = ":https"
		}
	}

	// Create net server
	if err := this.netserver(this.Config.Addr, tlsconfig, this.Config.Timeout); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Return success
	return this
}

func (this *server) Run(ctx context.Context, _ Provider) error {
	var result error
	go func() {
		<-ctx.Done()
		if err := this.stop(); err != nil {
			result = multierror.Append(result, err)
		}
	}()
	if err := this.runInForeground(); err != nil && errors.Is(err, http.ErrServerClosed) == false {
		result = multierror.Append(result, err)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *server) String() string {
	str := "<httpserver"
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
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MODULE

func Name() string {
	return "httpserver"
}

func GetConfig() interface{} {
	return &Config{}
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Serve requests over HTTP or through a unix socket\n")
	fmt.Fprintln(w, "  Configuration:")
	fmt.Fprintln(w, "    addr: <string>")
	fmt.Fprintln(w, "      Optional, listening network address or path for binding HTTP server")
	fmt.Fprintln(w, "      can be overridden by -addr argument on command line")
	fmt.Fprintln(w, "    timeout: <duration>")
	fmt.Fprintln(w, "      Optional, read timeout on requests over network")
	fmt.Fprintln(w, "    key: <path>")
	fmt.Fprintln(w, "      Optional, location of TLS key file when serving over network")
	fmt.Fprintln(w, "    cert: <path>")
	fmt.Fprintln(w, "      Required if key is set, location of TLS certificate file when serving over network")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *server) fcgiserver(path string) error {
	// Create server
	this.fcgi = &fcgi.Server{}
	this.fcgi.Handler = this.Router.(http.Handler)
	this.fcgi.Network = "unix"
	this.fcgi.Addr = path

	// Return success
	return nil
}

func (this *server) netserver(addr string, config *tls.Config, timeout time.Duration) error {
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
	this.srv.Handler = this.Router.(http.Handler)
	this.srv.ReadHeaderTimeout = timeout
	this.srv.IdleTimeout = timeout

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// START AND STOP

func (this *server) runInForeground() error {
	if this.fcgi != nil {
		return this.fcgi.ListenAndServe()
	} else if this.srv.TLSConfig != nil {
		return this.srv.ListenAndServeTLS("", "")
	} else {
		return this.srv.ListenAndServe()
	}
}

func (this *server) stop() error {
	if this.fcgi != nil {
		return this.fcgi.Close()
	} else {
		return this.srv.Close()
	}
}
