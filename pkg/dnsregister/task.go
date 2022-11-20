package dnsregister

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	// Package imports
	event "github.com/mutablelogic/go-server/pkg/event"
	task "github.com/mutablelogic/go-server/pkg/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	task.Task
	*Register

	delta   time.Duration
	modtime map[string]time.Time // Last time each hostname was registered
	records []Record             // Set of records to register
	addr    net.IP               // Current external address
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	this := &t{
		Register: New(p.Timeout()),
		modtime:  make(map[string]time.Time, len(p.Records_)+1),
		delta:    p.Delta(),
		records:  p.Records(),
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
	for _, r := range t.records {
		str += fmt.Sprint(" ", r)
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
			ticker.Reset(10 * time.Second)
			if t.touch("*") {
				if changed, err := t.external(); err != nil {
					t.Emit(event.Error(ctx, err))
				} else if changed {
					t.Emit(event.Infof(ctx, ExternalIP, "External address changed to %q", t.addr))
				} else {
					t.Emit(event.Infof(ctx, NotModified, "External address not modified"))
				}
			} else if record := t.next(); !record.IsZero() {
				if addr, err := t.register(record); errors.Is(err, ErrNotModified) {
					t.Emit(event.Infof(ctx, NotModified, "Not modified: %q", record.Name))
				} else if err != nil {
					t.Emit(event.Error(ctx, err))
				} else {
					t.Emit(event.Infof(ctx, Modified, "Registered: %q => %q", record.Name, addr))
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

// return an entry from the register, which has not been updated recently
func (t *t) next() Record {
	for _, record := range t.records {
		if record.Name == "" || record.Name == "*" {
			continue
		}
		if t.touch(record.Name) {
			return record
		}
	}
	// Return empty record
	return Record{}
}

// get a new external address
func (t *t) external() (bool, error) {
	addr, err := t.GetExternalAddress()
	if err != nil {
		t.addr = nil
		return false, err
	}

	// Return if address has not changed
	if t.addr.String() == addr.String() {
		return false, nil
	}

	// Update address
	t.addr = addr

	// Re-register all hosts
	for k := range t.modtime {
		if k != "*" {
			delete(t.modtime, k)
		}
	}

	// Return success
	return true, nil
}

// get a new external address
func (t *t) register(r Record) (net.IP, error) {
	// Check parameters
	if r.IsZero() {
		return nil, ErrBadParameter.With("Invalid record")
	}

	// If record address is empty, use external address
	if r.Address == "" {
		r.Address = t.addr.String()
	}

	// Register the address
	if ip := r.IP(); ip == nil {
		return nil, ErrBadParameter.Withf("Invalid address: %q", r.Address)
	} else if err := t.RegisterAddress(r.Name, r.User, r.Password, ip, false); err != nil {
		return ip, fmt.Errorf("RegisterAddress: %q: %w", ip, err)
	} else {
		return ip, nil
	}
}
