package httpserver

import (
	"context"
	"crypto/tls"
	"time"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/router"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Router_  types.Task     `json:"router,omitempty"` // The router object which serves the gateways
	Listen_  types.String   `json:"listen,omitempty"` // Address or path for binding HTTP server
	TLS_     *TLS           `json:"tls,omitempty"`    // TLS parameters, or nil if not using TLS
	Timeout_ types.Duration `hcl:"timeout,optional"`  // Read timeout on HTTP requests

	router  plugin.Router
	listen  string
	tls     *tls.Config
	timeout time.Duration
}

type TLS struct {
	Key_  types.String `json:"key"`  // Path to TLS Private Key
	Cert_ types.String `json:"cert"` // Path to TLS Certificate
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName      = "httpserver"
	defaultListen    = ":http"
	defaultListenTLS = ":https"
	defaultTimeout   = types.Duration(10 * time.Second)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	p.router = p.Router(ctx, provider)
	p.listen = p.Listen()
	p.timeout = p.Timeout()
	if tls, err := p.TLS(); err != nil {
		return nil, err
	} else {
		p.tls = tls
	}

	return NewWithPlugin(p)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

func (p Plugin) Listen() string {
	if p.Listen_ == "" {
		if p.TLS_ == nil {
			return defaultListen
		} else {
			return defaultListenTLS
		}
	} else {
		return string(p.Listen_)
	}
}

func (p Plugin) TLS() (*tls.Config, error) {
	if p.TLS_ == nil {
		return nil, nil
	}
	if cert, err := tls.LoadX509KeyPair(string(p.TLS_.Cert_), string(p.TLS_.Key_)); err != nil {
		return nil, err
	} else {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}
}

func (p Plugin) Router(ctx context.Context, provider iface.Provider) plugin.Router {
	if p.Router_.Task == nil {
		plugin := router.Plugin{}
		if router, err := provider.New(ctx, plugin); err != nil {
			// We currently panic if we get here, it's not expected
			panic(err)
		} else {
			p.Router_.Task = router
		}
	}
	if router, ok := p.Router_.Task.(plugin.Router); ok && router != nil {
		return router
	} else {
		return nil
	}
}

func (p Plugin) Timeout() time.Duration {
	if p.Timeout_ <= 0 {
		return time.Duration(defaultTimeout)
	} else {
		return time.Duration(p.Timeout_)
	}
}
