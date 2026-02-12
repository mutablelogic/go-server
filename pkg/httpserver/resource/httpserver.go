package resource

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Resource struct {
	Listen       string                  `name:"listen" help:"Listen address (e.g. localhost:8080)"`
	Description  string                  `name:"description" help:"Server description for OpenAPI spec"`
	Endpoint     string                  `name:"endpoint" readonly:"" help:"Base URL of the running server"`
	Router       schema.ResourceInstance `name:"router" type:"httprouter" required:"" help:"HTTP router"`
	ReadTimeout  time.Duration           `name:"read-timeout" default:"5m" help:"Read timeout"`
	WriteTimeout time.Duration           `name:"write-timeout" default:"5m" help:"Write timeout"`
	IdleTimeout  time.Duration           `name:"idle-timeout" default:"5m" help:"Idle timeout for keep-alive connections"`
	TLS          struct {
		Name   string `name:"name" help:"TLS server name"`
		Verify bool   `name:"verify" default:"true" help:"Verify client certificates"`
		Cert   []byte `name:"cert" sensitive:"" help:"TLS certificate PEM data"`
		Key    []byte `name:"key" sensitive:"" help:"TLS key PEM data"`
	} `embed:"" prefix:"tls."`
}

type ResourceInstance struct {
	provider.ResourceInstance[Resource]

	// runtime state
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	runtimeErr atomic.Pointer[error] // error from the background server goroutine
}

var _ schema.Resource = (*Resource)(nil)
var _ schema.ResourceInstance = (*ResourceInstance)(nil)
var _ server.HTTPServer = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - HTTP SERVER

// Spec returns the OpenAPI server entry for this instance, or nil if
// the server has not started yet.
func (r *ResourceInstance) Spec() *openapi.Server {
	c := r.State()
	if c == nil || c.Endpoint == "" {
		return nil
	}
	return &openapi.Server{URL: c.Endpoint, Description: c.Description}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r Resource) New(name string) (schema.ResourceInstance, error) {
	return &ResourceInstance{
		ResourceInstance: provider.NewResourceInstance[Resource](r, name),
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
	v, err := r.ResourceInstance.Validate(ctx, state, resolve)
	if err != nil {
		return nil, err
	}
	desired := v.(*Resource)

	// Timeouts must be non-negative
	if desired.ReadTimeout < 0 {
		return nil, httpresponse.ErrBadRequest.With("negative read timeout")
	}
	if desired.WriteTimeout < 0 {
		return nil, httpresponse.ErrBadRequest.With("negative write timeout")
	}
	if desired.IdleTimeout < 0 {
		return nil, httpresponse.ErrBadRequest.With("negative idle timeout")
	}

	return desired, nil
}

// Plan computes the diff and performs pre-flight checks on changed
// fields: probing the listen address for availability and validating
// TLS certificates (PEM parsing, key-pair match, expiry).
func (r *ResourceInstance) Plan(ctx context.Context, v any) (schema.Plan, error) {
	plan, err := r.ResourceInstance.Plan(ctx, v)
	if err != nil {
		return plan, err
	}

	// Scan the changes for fields we need to pre-flight check
	var newAddr string
	var tlsChanging bool
	for _, ch := range plan.Changes {
		switch ch.Field {
		case "listen":
			if s, ok := ch.New.(string); ok {
				newAddr = s
			}
		case "tls.cert", "tls.key":
			tlsChanging = true
		}
	}

	// Probe the listen address if it's changing
	if newAddr != "" {
		addr := httpserver.ListenAddr(newAddr, false)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return plan, httpresponse.ErrBadRequest.Withf("listen: %v", err)
		}
		ln.Close()
	}

	// Validate TLS certificate and key if either is changing
	if tlsChanging {
		desired, ok := v.(*Resource)
		if ok && (len(desired.TLS.Cert) > 0 || len(desired.TLS.Key) > 0) {
			if err := httpserver.ValidateCert(desired.TLS.Cert, desired.TLS.Key); err != nil {
				return plan, httpresponse.ErrBadRequest.Withf("tls: %v", err)
			}
		}
	}

	return plan, nil
}

// Apply materialises the resource using the validated configuration.
func (r *ResourceInstance) Apply(ctx context.Context, v any) error {
	return r.ApplyConfig(ctx, v, func(ctx context.Context, c *Resource) error {
		// Stop the existing server if running
		if err := r.stop(); err != nil {
			return err
		}

		// The router must implement http.Handler
		router, ok := c.Router.(interface {
			ServeHTTP(http.ResponseWriter, *http.Request)
		})
		if !ok {
			return httpresponse.ErrBadRequest.With("router does not implement http.Handler")
		}

		// Build TLS config if any TLS data is provided
		var cert *tls.Config
		if len(c.TLS.Cert) > 0 || len(c.TLS.Key) > 0 {
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
			httpserver.WithIdleTimeout(c.IdleTimeout),
		)
		if err != nil {
			return err
		}

		// Bind the port synchronously so callers get an immediate error
		// (e.g. "address already in use") instead of a silent goroutine failure.
		if err := srv.Listen(); err != nil {
			return err
		}

		// Compute the endpoint URL from the actual bound address
		scheme := "http"
		if cert != nil {
			scheme = "https"
		}
		c.Endpoint = scheme + "://" + srv.Addr()

		// Run the server in the background – use Background() as the parent
		// so the server outlives the HTTP request that triggered Apply.
		srvCtx, cancel := context.WithCancel(context.Background())
		r.cancel = cancel
		r.runtimeErr.Store(nil)
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			if err := srv.Run(srvCtx); err != nil && !errors.Is(err, context.Canceled) {
				r.runtimeErr.Store(&err)
			}
		}()

		return nil
	})
}

// Destroy tears down the resource and releases its backing
// infrastructure. It returns an error if the resource cannot be
// cleanly removed.
func (r *ResourceInstance) Destroy(_ context.Context) error {
	return r.stop()
}

// RuntimeErr returns the error from the background server goroutine,
// or nil if the server is running normally.
func (r *ResourceInstance) RuntimeErr() error {
	if p := r.runtimeErr.Load(); p != nil {
		return *p
	}
	return nil
}

// Read returns the current state of the server instance. If the
// background server has exited with an error, it is returned.
func (r *ResourceInstance) Read(ctx context.Context) (schema.State, error) {
	state, err := r.ResourceInstance.Read(ctx)
	if err != nil {
		return state, err
	}
	if runtimeErr := r.RuntimeErr(); runtimeErr != nil {
		return state, runtimeErr
	}
	return state, nil
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
