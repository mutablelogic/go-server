package dnsregister

import (
	"context"
	"fmt"
	"net"
	"time"

	// Package imports
	task "github.com/mutablelogic/go-server/pkg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	task.Task
	*Register

	delta   time.Duration
	modtime map[string]time.Time // Last time each hostname was registered
	addr    net.IP               // Current external address
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	this := &t{
		Register: New(p.Timeout()),
		modtime:  make(map[string]time.Time, len(p.Records)+1),
		delta:    p.Delta(),
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *t) String() string {
	str := "<dnsregister"
	str += fmt.Sprint(" ", t.Register)
	if t.delta > 0 {
		str += fmt.Sprint(" delta=", t.delta)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Run(ctx context.Context) error {
	// Create ticker for doing the registration
	ticker := time.NewTimer(time.Second)
	defer ticker.Stop()

	// Run loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Reset(30 * time.Second)
			if t.touch("*") {
				fmt.Println("Get external address")
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// touch the modtime for the given hostname to see if we should re-register
func (t *t) touch(key string) bool {
	modtime, exists := t.modtime[key]
	if !exists || modtime.IsZero() {
		t.modtime[key] = time.Now()
		return true
	}
	if time.Since(modtime) >= t.delta {
		t.modtime[key] = time.Now()
		return true
	}
	return false
}
