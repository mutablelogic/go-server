package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ref "github.com/mutablelogic/go-server/pkg/ref"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type server struct {
	http http.Server
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
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(listen string, router http.Handler, cert *tls.Config, opts ...Opt) (*server, error) {
	server := new(server)

	// Apply options
	opt, err := apply(opts...)
	if err != nil {
		return nil, err
	}

	// Set default listen host
	if listen == "" {
		listen = defaultListenHost
	}

	// Join host and port
	if !strings.Contains(listen, ":") {
		if cert == nil {
			listen = net.JoinHostPort(listen, defaultListenPortHttp)
		} else {
			listen = net.JoinHostPort(listen, defaultListenPortHttps)
		}
	}

	// Set other options
	server.http.Addr = listen
	server.http.TLSConfig = cert
	server.http.Handler = router
	server.http.WriteTimeout = opt.w
	server.http.ReadTimeout = opt.r

	// Return success
	return server, nil
}

///////////////////////////////////////////////////////////////////////////////
// TASK

func (server *server) Run(parent context.Context) error {
	var wg sync.WaitGroup

	// Make a cancelable context
	ctx, cancel := context.WithCancel(parent)

	// Wait for cancel in a background process
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		// Wait for context to be done
		<-ctx.Done()

		// Stop server, terminate connections after 30 seconds
		if err := server.shutdown(); err != nil {
			ref.Log(parent).Print(parent, err)
		}
	}(ctx)

	// Run server as a foreground process
	var result error
	if server.http.TLSConfig != nil {
		result = server.http.ListenAndServeTLS("", "")
	} else {
		result = server.http.ListenAndServe()
	}

	// Cancel waiting for CTRL+C, if not already cancelled
	cancel()

	// Wait for the goroutine to complete
	wg.Wait()

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
