package mdns

import (
	"context"
	"net"
	"time"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Interface string `json:"interface,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultName  = "mdns"
	DefaultLabel = "local"
)

var (
	multicastAddrIp4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	multicastAddrIp6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
	sendRetryCount   = 3
	sendRetryDelta   = 100 * time.Millisecond
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(_ context.Context, _ iface.Provider) (iface.Task, error) {
	if err := p.HasNameLabel(); err != nil {
		return nil, err
	}

	// Return new task from plugin configuration
	return NewWithPlugin(p)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithLabel(label string) Plugin {
	return Plugin{
		Plugin: task.WithLabel(DefaultName, label),
	}
}

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return DefaultName
	}
}

func (p Plugin) Label() string {
	if label := p.Plugin.Label(); label != "" {
		return label
	} else {
		return DefaultLabel
	}
}

func (p Plugin) Interfaces() ([]net.Interface, error) {
	// Check interfaces
	var i net.Interface
	var err error
	if p.Interface != "" {
		i, err = interfaceForName(p.Interface)
		if err != nil {
			return nil, err
		}
	}
	if ifaces, err := multicastInterfaces(i); err != nil {
		return nil, err
	} else if len(ifaces) == 0 {
		return nil, ErrBadParameter.With("no multicast interfaces defined")
	} else {
		return ifaces, nil
	}
}
