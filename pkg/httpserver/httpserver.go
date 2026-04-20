package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type server struct {
	http       http.Server
	listener   net.Listener
	mux        *http.ServeMux
	serverName string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultListenHost      = "localhost"
	defaultListenPortHttp  = "http"
	defaultListenPortHttps = "https"
	defaultShutdownTimeout = 30 * time.Second
	defaultReadTimeout     = 5 * time.Minute
	defaultWriteTimeout    = 5 * time.Minute
	defaultIdleTimeout     = 5 * time.Minute
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// ListenAddr resolves a listen address string into a host:port pair,
// applying defaults for missing host or port. When tls is true the
// default port is "https" (443); otherwise it is "http" (80).
func ListenAddr(listen string, tls bool) string {
	if listen == "" {
		listen = defaultListenHost
	}
	if !strings.Contains(listen, ":") {
		if tls {
			listen = net.JoinHostPort(listen, defaultListenPortHttps)
		} else {
			listen = net.JoinHostPort(listen, defaultListenPortHttp)
		}
	}
	return listen
}

func New(listen string, tls *tls.Config, opts ...Opt) (*server, error) {
	server := new(server)

	// Apply options
	opt, err := apply(opts...)
	if err != nil {
		return nil, err
	}

	// Resolve the listen address
	listen = ListenAddr(listen, tls != nil && len(tls.Certificates) > 0)

	// Set other options
	server.mux = http.NewServeMux()
	server.http.Handler = server.mux
	server.http.Addr = listen
	if tls != nil && len(tls.Certificates) > 0 {
		server.http.TLSConfig = tls
	}
	server.http.WriteTimeout = opt.w
	server.http.ReadTimeout = opt.r
	server.http.IdleTimeout = opt.i

	// Set server name from TLS config if available (used for telemetry, OpenAPI and auth issuer certs)
	if tls != nil && tls.ServerName != "" {
		server.serverName = strings.TrimSpace(tls.ServerName)
	}

	// Return success
	return server, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Router returns the server's http.ServeMux for registering handlers.
func (server *server) Router() *http.ServeMux {
	return server.mux
}

// Addr returns the listen address. After [Listen] has been called this
// returns the actual bound address (which may differ from the configured
// address when an ephemeral port is used).
func (server *server) Addr() string {
	if server.listener != nil {
		return server.listener.Addr().String()
	}
	return server.http.Addr
}

// Listen binds the server to its configured address. It returns
// immediately with an error if the port is already in use. After a
// successful call, use [Run] to start serving requests.
func (server *server) Listen() error {
	ln, err := net.Listen("tcp", server.http.Addr)
	if err != nil {
		return err
	}
	// Wrap with TLS if configured
	if server.http.TLSConfig != nil {
		ln = tls.NewListener(ln, server.http.TLSConfig)
	}
	server.listener = ln
	return nil
}

func (server *server) URL() *url.URL {
	var url url.URL
	url.Scheme = defaultListenPortHttp
	url.Path = "/"
	url.Host = server.Addr()
	if server.serverName != "" {
		url.Scheme = defaultListenPortHttps
		url.Host = server.serverName
	}
	return types.Ptr(url)
}

///////////////////////////////////////////////////////////////////////////////
// TASK

// Run serves requests on the listener obtained from a prior call to
// [Listen]. If Listen has not been called, Run calls it internally
// (preserving backward compatibility). Run blocks until the context is
// cancelled, then gracefully shuts down.
func (server *server) Run(parent context.Context) error {
	// If Listen was not called beforehand, call it now for backward compat
	if server.listener == nil {
		if err := server.Listen(); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	// Make a cancelable context
	ctx, cancel := context.WithCancel(parent)

	// Wait for cancel in a background process — capture the shutdown
	// error in its own variable to avoid a data race with Serve.
	var shutdownErr error
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		// Wait for context to be done
		<-ctx.Done()

		// Stop server, terminate connections after 30 seconds
		shutdownErr = server.shutdown()
	}(ctx)

	// Serve on the already-bound listener
	serveErr := server.http.Serve(server.listener)

	// Cancel waiting for CTRL+C, if not already cancelled
	cancel()

	// Wait for the shutdown goroutine to complete
	wg.Wait()

	// Join errors — safe now because both goroutines have finished
	result := errors.Join(serveErr, shutdownErr)

	// If result is http.ErrServerClosed, return nil
	if errors.Is(result, http.ErrServerClosed) {
		return nil
	} else {
		return result
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (server *server) shutdown() error {
	shutdownContext, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	if err := server.http.Shutdown(shutdownContext); errors.Is(err, context.DeadlineExceeded) {
		return httpresponse.ErrInternalError.With("shutdown timeout, forced connections to close")
	} else {
		return err
	}
}
