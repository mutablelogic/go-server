package httpserver

import (
	"context"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/terraform-provider-nginx"

	// Module imports
	"github.com/mutablelogic/terraform-provider-nginx/pkg/router"
	util "github.com/mutablelogic/terraform-provider-nginx/pkg/util"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Label   string        `hcl:"label,label"`
	Router  Task          `hcl:"router,optional"`
	Addr    string        `hcl:"listen,optional"`  // Address or path for binding HTTP server
	TLS     *TLS          `hcl:"tls,block"`        // TLS parameters
	Timeout time.Duration `hcl:"timeout,optional"` // Read timeout on HTTP requests
}

type TLS struct {
	Key  string `hcl:"key"`  // Path to TLS Private Key
	Cert string `hcl:"cert"` // Path to TLS Certificate
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultLabel   = "httpserver"
	DefaultTimeout = 10 * time.Second
)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) Name() string {
	return DefaultLabel
}

// Return a new task. Label for the task can be retrieved from context
func (c Config) New(ctx context.Context, provider Provider) (Task, error) {
	// Set label
	if c.Label == "" {
		c.Label = DefaultLabel
	}
	// Set timeout
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	// Check label
	if !util.IsIdentifier(c.Label) {
		return nil, ErrBadParameter.Withf("label: %q", c.Label)
	}
	// Create a router if it's not provided
	if c.Router == nil {
		if router, err := provider.New(ctx, router.Config{
			Label: c.Label + "-router",
		}); err != nil {
			return nil, err
		} else {
			c.Router = router
		}
	}

	// Return configuration
	return NewWithConfig(c)
}
