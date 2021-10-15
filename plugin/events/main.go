package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Capacity int `yaml:"capacity"`
}

type plugin struct {
	sync.Mutex
	Config
	C chan Event
	S map[string]chan<- Event
	E []event
}

type event struct {
	Event
	ts    time.Time
	count int64
}

type Counter int64

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	queueCapacity        = 100
	defaultEventCapacity = 100
	maxEventCapacity     = 1000
)

var (
	counter = new(Counter)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the eventbus module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Read configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		p.Config = cfg
	}

	// Set event capacity
	if p.Capacity <= 0 {
		p.Capacity = defaultEventCapacity
	} else {
		p.Capacity = minInt(p.Capacity, maxEventCapacity)
	}

	// Create receive and send channel
	p.C = make(chan Event, queueCapacity)
	p.S = make(map[string]chan<- Event)
	p.E = make([]event, 0, p.Capacity)

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<events"
	str += fmt.Sprint(" capacity=", p.Capacity)
	if len(p.S) > 0 {
		n := make([]string, 0, len(p.S))
		for name := range p.S {
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
	fmt.Fprintln(w, "\n  Configuration:")
	fmt.Fprintln(w, "    capacity: <number>")
	fmt.Fprintln(w, "      Optional, number of historical events to store")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "events"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Add handlers
	if err := p.AddHandlers(ctx, provider); err != nil {
		return err
	}

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-p.C:
			for name, s := range p.S {
				select {
				case s <- evt:
					// no-op
				default:
					provider.Printf(ctx, "cannot send on blocked channel: %q", name)
				}
			}
			p.Lock()
			p.E = append(p.E, event{evt, time.Now(), counter.Next()})
			p.Unlock()
		case <-ticker.C:
			// Reduce size of event buffer when it reaches capacity
			if len(p.E) > p.Capacity {
				cull := len(p.E) - p.Capacity
				p.Lock()
				p.E = p.E[cull:]
				p.Unlock()
			}
		}
	}

	// Release resources
	close(p.C)
	p.S = nil

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Post(ctx context.Context, evt Event) {
	select {
	case p.C <- evt:
		break
	default:
		panic("events: cannot post on blocked channel")
	}
}

func (p *plugin) Subscribe(ctx context.Context, s chan<- Event) error {
	if name := provider.ContextPluginName(ctx); name == "" {
		return ErrBadParameter.With("Subscribe")
	} else if _, exists := p.S[name]; exists {
		return ErrDuplicateEntry.With("Subscribe ", strconv.Quote(name))
	} else {
		p.S[name] = s
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return next counter value, which cycles positive values from 1 to MaxInt64
func (c *Counter) Next() int64 {
	n := atomic.AddInt64((*int64)(c), 1)
	if n == math.MaxInt64 {
		atomic.StoreInt64((*int64)(c), 1)
		return 1
	} else {
		return n
	}
}
