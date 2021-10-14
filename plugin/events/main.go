package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
	C chan Event
	S map[string]chan<- Event
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 100
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the eventbus module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Create receive and send channel
	this.C = make(chan Event, defaultCapacity)
	this.S = make(map[string]chan<- Event)

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<events"
	if len(this.S) > 0 {
		n := make([]string, 0, len(this.S))
		for name := range this.S {
			n = append(n, name)
		}
		str += fmt.Sprintf(" subscribers=%q", n)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// USAGE

func Usage(w io.Writer) {
	fmt.Fprintln(w, "\n  Publish and subscribe to plugin events.")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "events"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
FOR_LOOP:
	for {
		select {
		case evt := <-this.C:
			if evt.Value() == nil {
				provider.Printf(ctx, "events: %q", evt.Name())
			} else {
				provider.Printf(ctx, "events: %q: %v", evt.Name(), evt.Value())
			}
			for name, s := range this.S {
				select {
				case s <- evt:
					// no-op
				default:
					provider.Printf(ctx, "events: cannot send on blocked channel: %q", name)
				}
			}
		case <-ctx.Done():
			break FOR_LOOP
		}
	}

	// Release resources
	close(this.C)
	this.S = nil

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - EVENT BUS

func (this *plugin) Post(ctx context.Context, evt Event) {
	select {
	case this.C <- evt:
		break
	default:
		panic("events: cannot post on blocked channel")
	}
}

func (this *plugin) Subscribe(ctx context.Context, s chan<- Event) error {
	if name := provider.ContextPluginName(ctx); name == "" {
		return ErrBadParameter.With("Subscribe")
	} else if _, exists := this.S[name]; exists {
		return ErrDuplicateEntry.With("Subscribe ", strconv.Quote(name))
	} else {
		this.S[name] = s
	}

	// Return success
	return nil
}
