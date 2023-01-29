package main

import (
	"context"
	"net/url"
	"time"

	// Packages
	pool "github.com/mutablelogic/go-accessory/pkg/pool"
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	"github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Url_      string         `json:"url,omitempty"`      // Database URL
	Database_ string         `json:"database,omitempty"` // Database Name
	Limit_    uint           `json:"limit,omitempty"`    // Concurrent Connection limit
	Timeout_  types.Duration `json:"timeout,omitempty"`  // Database timeout
	Trace_    bool           `json:"trace,omitempty"`    // Trace database calls
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "accessory-connpool"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new connection pool task from plugin configuration
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	} else {
		return NewWithPlugin(ctx, p)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a plugin with a specific label
func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(defaultName, label),
	}
}

// Return the name of the plugin
func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

// Set the URL for the plugin
func (p Plugin) WithUrl(url string) Plugin {
	p.Url_ = url
	return p
}

// Return the URL for the plugin
func (p Plugin) Url() (*url.URL, error) {
	if p.Url_ == "" {
		return nil, ErrBadParameter.With("Url")
	} else if url, err := url.Parse(p.Url_); err != nil {
		return nil, err
	} else {
		return url, nil
	}
}

// Return the pool options
func (p Plugin) Options() []pool.Option {
	var opts []pool.Option
	if p.Name_ != "" {
		opts = append(opts, pool.OptDatabase(p.Database_))
	}
	if p.Limit_ > 0 {
		opts = append(opts, pool.OptMaxSize(int64(p.Limit_)))
	}
	if p.Timeout_ > 0 {
		opts = append(opts, pool.OptTimeout(time.Duration(p.Timeout_)))
	}
	return opts
}

// Return true if tracing is enabled
func (p Plugin) Trace() bool {
	return p.Trace_
}
