package dnsregister

import (
	"context"
	"fmt"
	"net"
	"time"

	// Package imports

	"github.com/mutablelogic/go-server/pkg/event"
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
				if err := t.register(); err != nil {
					t.Emit(event.Error(ctx, err))
				}
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

// return an entry from the register
func (t *t) next() Record {
	return Record{}
	/*
	   	for _, record := range t.records {
	   		if p.touch(record.) && p.addr != nil {
	   			return host, user, passwd
	   		}
	   	}

	   // No entry found
	   return Record{}
	*/
}

// get a new external address
func (t *t) register() error {
	addr, err := t.GetExternalAddress()
	if err != nil {
		t.addr = nil
		return err
	}

	// Return if address has not changed
	if t.addr.String() == addr.String() {
		return nil
	}

	// Update address
	t.addr = addr

	fmt.Println("Registering", addr)

	// Re-register all hosts
	for k := range t.modtime {
		if k != "*" {
			delete(t.modtime, k)
		}
	}

	// Return success
	return nil
}
