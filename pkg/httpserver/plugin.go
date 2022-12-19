package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	router "github.com/mutablelogic/go-server/pkg/httpserver/router"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Router_  types.Task     `json:"router,omitempty"`  // The router object which serves the gateways
	Listen_  types.String   `json:"listen,omitempty"`  // Address or path for binding HTTP server
	TLS_     *TLS           `json:"tls,omitempty"`     // TLS parameters, or nil if not using TLS (ignored for file sockets)
	Timeout_ types.Duration `json:"timeout,omitempty"` // Read timeout on HTTP requests (ignored for file sockets)
	Owner_   types.String   `json:"owner,omitempty"`   // Owner of the socket file (ignored for network sockets)
	Group_   types.String   `json:"group,omitempty"`   // Owner Group of the socket file (ignored for network sockets)
	Routes   []router.Route `json:"routes"`            // Array of routes, required
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
	defaultMode      = os.FileMode(0600)
	groupMode        = os.FileMode(0060)
	allMode          = os.FileMode(0777)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Create the task and return and return any errors
	return NewWithPlugin(p, p.Router(ctx, provider))
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
	if p.TLS_.Cert_ == "" && p.TLS_.Key_ == "" {
		return nil, nil
	}
	if cert, err := tls.LoadX509KeyPair(string(p.TLS_.Cert_), string(p.TLS_.Key_)); err != nil {
		return nil, fmt.Errorf("LoadX509KeyPair: %w", err)
	} else {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}
}

func (p Plugin) Router(ctx context.Context, provider iface.Provider) plugin.Router {
	if p.Router_.Task == nil {
		plugin := router.WithLabel(p.Label()).WithRoutes(p.Routes)
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

func (p Plugin) Owner() (int, error) {
	if p.Owner_ == "" {
		return -1, nil
	}
	if user, err := user.Lookup(string(p.Owner_)); err != nil {
		return -1, err
	} else if uid, err := strconv.ParseUint(user.Uid, 0, 32); err != nil {
		return -1, err
	} else {
		return int(uid), nil
	}
}

func (p Plugin) Group() (int, error) {
	if p.Group_ == "" {
		return -1, nil
	}
	if group, err := user.LookupGroup(string(p.Group_)); err != nil {
		return -1, err
	} else if gid, err := strconv.ParseUint(group.Gid, 0, 32); err != nil {
		return -1, err
	} else {
		return int(gid), nil
	}
}

func (p Plugin) Mode() os.FileMode {
	mode := defaultMode
	if gid, err := p.Group(); gid != -1 && err == nil {
		mode |= groupMode
	}
	return mode & allMode
}
