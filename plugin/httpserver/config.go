package httpserver

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Listen       *url.URL      `kong:"-"`
	Router       string        `kong:"-"`
	ReadTimeout  time.Duration `name:"read-timeout" default:"5m" help:"Read timeout"`
	WriteTimeout time.Duration `name:"write-timeout" default:"5m" help:"Write timeout"`
	TLS          struct {
		Name   string `name:"name" default:"${HOST}" help:"TLS server name"`
		Verify bool   `name:"verify" default:"true" help:"Verify client certificates"`
		Cert   string `name:"cert" type:"file" default:"" help:"TLS certificate PEM file"`
		Key    string `name:"key" type:"file" default:"" help:"TLS key PEM file"`
	} `embed:"" prefix:"tls."`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	// Get the router from the provider
	router, ok := provider.Provider(ctx).Task(ctx, c.Router).(http.Handler)
	if !ok || router == nil {
		return nil, httpresponse.ErrNotFound.Withf("Router %q not found", c.Router)
	}

	// Check listen scheme
	if c.Listen.Scheme != types.SchemeInsecure && c.Listen.Scheme != types.SchemeSecure {
		return nil, httpresponse.ErrBadRequest.Withf("Invalid listen scheme: %q", c.Listen.Scheme)
	}

	// Get the TLS Config
	var tls *tls.Config
	if c.Listen.Scheme == types.SchemeSecure {
		if tls_, err := httpserver.TLSConfig(c.TLS.Name, c.TLS.Verify, c.TLS.Cert, c.TLS.Key); err != nil {
			return nil, err
		} else {
			tls = tls_
		}
	}

	// Check that TLS certificate and key are provided for secure server only
	if c.Listen.Scheme == types.SchemeInsecure {
		if c.TLS.Cert != "" || c.TLS.Key != "" {
			return nil, httpresponse.ErrBadRequest.Withf("TLS certificate or key not required for insecure server")
		}
	}

	// Return the server
	return httpserver.New(c.Listen.Host, router, tls, httpserver.WithReadTimeout(c.ReadTimeout), httpserver.WithWriteTimeout(c.WriteTimeout))
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return "httpserver"
}

func (c Config) Description() string {
	return "HTTP server"
}
