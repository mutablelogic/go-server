package dnsregister

import (
	"context"
	"time"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Timeout_ types.Duration `json:"timeout,omitempty"` // Read timeout on HTTP requests
	Delta_   types.Duration `json:"delta,omitempty"`   // Time between updating information
	Records  []Record       `json:"records"`           // Records to register
}

type Record struct {
	Name     string `json:"name"`               // Name of record
	Address  string `json:"address,omitempty"`  // Address of record
	User     string `json:"user,omitempty"`     // Username for registration
	Password string `json:"password,omitempty"` // Password for registration
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName      = "dnsregister"
	defaultTimeout   = types.Duration(10 * time.Second)
	userAgent        = "github.com/mutablelogic/go-server/pkg/dnsregister"
	apifyUrl         = "https://api.ipify.org"
	googleDomainsUrl = "https://domains.google.com/nic/update"
	ratePeriod       = 30 * time.Second // Maximum of two requests per minute
	minDelta         = 5 * time.Minute
	defaultDelta     = 30 * time.Minute
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(ctx context.Context, provider iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
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

func (p Plugin) Timeout() time.Duration {
	if p.Timeout_ <= 0 {
		return time.Duration(defaultTimeout)
	} else {
		return time.Duration(p.Timeout_)
	}
}

func (p Plugin) Delta() time.Duration {
	if p.Delta_ == 0 {
		return time.Duration(defaultDelta)
	} else if p.Delta_ <= types.Duration(minDelta) {
		return time.Duration(minDelta)
	} else {
		return time.Duration(p.Delta_)
	}
}
