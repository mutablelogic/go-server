package resource

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Resource struct {
	Listen       string                  `name:"listen" help:"Listen address (e.g. localhost:8080)"`
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
	name   string
	config Resource

	// runtime state
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var _ schema.Resource = (*Resource)(nil)
var _ schema.ResourceInstance = (*ResourceInstance)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	counter = atomic.Int64{}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (r Resource) New() (schema.ResourceInstance, error) {
	return &ResourceInstance{
		name:   fmt.Sprintf("%s-%02d", r.Name(), counter.Add(1)),
		config: r,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE

func (Resource) Name() string {
	return "httpserver"
}

func (Resource) Schema() []schema.Attribute {
	return schema.Attributes(Resource{})
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - RESOURCE INSTANCE

// Return a unique name for this resource instance
func (r *ResourceInstance) Name() string {
	return r.name
}

// Resource returns the resource type that created this instance.
func (r *ResourceInstance) Resource() schema.Resource {
	return types.Ptr(r.config)
}

// Validate checks that the resource configuration is complete and
// internally consistent. It is called before Plan or Apply.
func (r *ResourceInstance) Validate(_ context.Context) error {
	c := r.config

	// Validate resource instance references (required + type constraints)
	if err := schema.ValidateRefs(c); err != nil {
		return httpresponse.ErrBadRequest.With(err.Error())
	}

	// Timeouts must be non-negative
	if c.ReadTimeout < 0 {
		return httpresponse.ErrBadRequest.With("negative read timeout")
	}
	if c.WriteTimeout < 0 {
		return httpresponse.ErrBadRequest.With("negative write timeout")
	}

	// TLS: cert and key must both be provided or both be absent
	if (c.TLS.Cert == "") != (c.TLS.Key == "") {
		return httpresponse.ErrBadRequest.With("tls.cert and tls.key must both be set or both be empty")
	}

	// TLS: if provided, cert and key files must be readable
	if c.TLS.Cert != "" {
		if _, err := os.Stat(c.TLS.Cert); err != nil {
			return httpresponse.ErrBadRequest.Withf("tls.cert: %v", err)
		}
	}
	if c.TLS.Key != "" {
		if _, err := os.Stat(c.TLS.Key); err != nil {
			return httpresponse.ErrBadRequest.Withf("tls.key: %v", err)
		}
	}

	return nil
}

// Plan computes the difference between the desired configuration and
// the current state, returning a set of planned changes without
// modifying anything. If current is nil the resource is being created.
func (r *ResourceInstance) Plan(_ context.Context, current schema.State) (schema.Plan, error) {
	desired := schema.StateOf(r.config)

	// No current state means this is a new resource
	if len(current) == 0 {
		changes := make([]schema.Change, 0, len(desired))
		for field, val := range desired {
			changes = append(changes, schema.Change{
				Field: field,
				New:   val,
			})
		}
		return schema.Plan{Action: schema.ActionCreate, Changes: changes}, nil
	}

	// Compare each desired field against the current state
	var changes []schema.Change
	for field, newVal := range desired {
		oldVal := current[field]
		if fmt.Sprint(oldVal) != fmt.Sprint(newVal) {
			changes = append(changes, schema.Change{
				Field: field,
				Old:   oldVal,
				New:   newVal,
			})
		}
	}

	if len(changes) == 0 {
		return schema.Plan{Action: schema.ActionNoop}, nil
	}
	return schema.Plan{Action: schema.ActionUpdate, Changes: changes}, nil
}

// Apply materialises the resource, creating or updating it to match
// the desired configuration. It returns the new state.
func (r *ResourceInstance) Apply(ctx context.Context, current schema.State) (schema.State, error) {
	c := r.config

	// Stop the existing server if running
	if err := r.stop(); err != nil {
		return nil, err
	}

	// The router must implement http.Handler
	router, ok := c.Router.(http.Handler)
	if !ok {
		return nil, httpresponse.ErrBadRequest.With("router does not implement http.Handler")
	}

	// Build TLS config if cert and key are provided
	var cert *tls.Config
	if c.TLS.Cert != "" && c.TLS.Key != "" {
		var err error
		cert, err = httpserver.TLSConfig(c.TLS.Name, c.TLS.Verify, c.TLS.Cert, c.TLS.Key)
		if err != nil {
			return nil, httpresponse.ErrBadRequest.Withf("tls: %v", err)
		}
	}

	// Create the HTTP server
	srv, err := httpserver.New(c.Listen, router, cert,
		httpserver.WithReadTimeout(c.ReadTimeout),
		httpserver.WithWriteTimeout(c.WriteTimeout),
	)
	if err != nil {
		return nil, err
	}

	// Run the server in the background
	srvCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := srv.Run(srvCtx); err != nil && !errors.Is(err, context.Canceled) {
			// TODO: surface runtime errors through the provider
			_ = err
		}
	}()

	return schema.StateOf(c), nil
}

// Destroy tears down the resource and releases its backing
// infrastructure. It returns an error if the resource cannot be
// cleanly removed.
func (r *ResourceInstance) Destroy(_ context.Context, current schema.State) error {
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

// References returns the labels of other resources this resource
// depends on. The runtime must ensure those resources are applied
// first and destroyed last.
func (r *ResourceInstance) References() []string {
	return schema.ReferencesOf(r.config)
}
