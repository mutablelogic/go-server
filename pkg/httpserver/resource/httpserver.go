package resource

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Resource struct {
	Listen       string                  `name:"listen" help:"Listen address (e.g. localhost:8080)"`
	Endpoint     string                  `name:"endpoint" readonly:"" help:"Base URL of the running server"`
	Router       schema.ResourceInstance `name:"router" type:"httprouter" required:"" help:"HTTP router"`
	ReadTimeout  time.Duration           `name:"read-timeout" default:"5m" help:"Read timeout"`
	WriteTimeout time.Duration           `name:"write-timeout" default:"5m" help:"Write timeout"`
	TLS          struct {
		Name   string `name:"name" help:"TLS server name"`
		Verify bool   `name:"verify" default:"true" help:"Verify client certificates"`
		Cert   string `name:"cert" type:"file" default:"" help:"TLS certificate PEM file"`
		Key    string `name:"key" type:"file" default:"" help:"TLS key PEM file"`
	} `embed:"" prefix:"tls."`
}

type ResourceInstance struct {
	provider.ResourceInstance[Resource]

	// runtime state
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var _ schema.Resource = (*Resource)(nil)
var _ schema.ResourceInstance = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r Resource) New() (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r),
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (Resource) Name() string {
	return "httpserver"
}

func (Resource) Schema() []schema.Attribute {
	return schema.AttributesOf(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// Validate decodes the incoming state, resolves references, and returns
// the validated *Resource configuration for use by Plan and Apply.
func (r *ResourceInstance) Validate(ctx context.Context, state schema.State, resolve schema.Resolver) (any, error) {
	desired, err := r.ResourceInstance.Validate(ctx, state, resolve)
	if err != nil {
		return nil, err
	}

	// Timeouts must be non-negative
	if desired.ReadTimeout < 0 {
		return nil, httpresponse.ErrBadRequest.With("negative read timeout")
	}
	if desired.WriteTimeout < 0 {
		return nil, httpresponse.ErrBadRequest.With("negative write timeout")
	}

	// TLS: cert and key must both be provided or both be absent
	if (desired.TLS.Cert == "") != (desired.TLS.Key == "") {
		return nil, httpresponse.ErrBadRequest.With("tls.cert and tls.key must both be set or both be empty")
	}

	// TLS: if provided, cert and key files must be readable
	if desired.TLS.Cert != "" {
		if _, err := os.Stat(desired.TLS.Cert); err != nil {
			return nil, httpresponse.ErrBadRequest.Withf("tls.cert: %v", err)
		}
	}
	if desired.TLS.Key != "" {
		if _, err := os.Stat(desired.TLS.Key); err != nil {
			return nil, httpresponse.ErrBadRequest.Withf("tls.key: %v", err)
		}
	}

	return desired, nil
}

// Apply materialises the resource using the validated configuration.
func (r *ResourceInstance) Apply(ctx context.Context, v any) error {
	c, ok := v.(*Resource)
	if !ok {
		return httpresponse.ErrInternalError.With("apply: unexpected config type")
	}

	// Stop the existing server if running
	if err := r.stop(); err != nil {
		return err
	}

	// The router must implement http.Handler
	router, ok := c.Router.(http.Handler)
	if !ok {
		return httpresponse.ErrBadRequest.With("router does not implement http.Handler")
	}

	// Build TLS config if cert and key are provided
	var cert *tls.Config
	if c.TLS.Cert != "" && c.TLS.Key != "" {
		var err error
		cert, err = httpserver.TLSConfig(c.TLS.Name, c.TLS.Verify, c.TLS.Cert, c.TLS.Key)
		if err != nil {
			return httpresponse.ErrBadRequest.Withf("tls: %v", err)
		}
	}

	// Create the HTTP server
	srv, err := httpserver.New(c.Listen, router, cert,
		httpserver.WithReadTimeout(c.ReadTimeout),
		httpserver.WithWriteTimeout(c.WriteTimeout),
	)
	if err != nil {
		return err
	}

	// Compute the endpoint URL
	scheme := "http"
	if cert != nil {
		scheme = "https"
	}
	c.Endpoint = scheme + "://" + srv.Addr()

	// Run the server in the background – use Background() as the parent
	// so the server outlives the HTTP request that triggered Apply.
	srvCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := srv.Run(srvCtx); err != nil && !errors.Is(err, context.Canceled) {
			// TODO: surface runtime errors through the provider
			_ = err
		}
	}()

	// Store the applied config and notify observers
	r.SetStateAndNotify(c, r)

	return nil
}

// Destroy tears down the resource and releases its backing
// infrastructure. It returns an error if the resource cannot be
// cleanly removed.
func (r *ResourceInstance) Destroy(_ context.Context) error {
	return r.stop()
}

// stop cancels the running server and waits for it to exit.
func (r *ResourceInstance) stop() error {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	r.wg.Wait()
	return nil
}
