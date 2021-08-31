package main

import (
	"context"
	"fmt"
	"strconv"

	// Modules
	. "github.com/djthorpe/go-server"
	"github.com/djthorpe/go-server/pkg/provider"
	sq "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Database string `yaml:"database"`
}

type plugin struct {
	sq.SQConnection
	schema string
	C      chan Event
	S      map[string]chan<- Event
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultDatabase = "main"
	defaultCapacity = 100
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the eventbus module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Create receive and send channel
	this.C = make(chan Event, defaultCapacity)
	this.S = make(map[string]chan<- Event)

	// Get sqlite
	if conn, ok := provider.GetPlugin(ctx, "sqlite").(sq.SQConnection); !ok {
		provider.Print(ctx, "missing sqlite dependency")
		return nil
	} else {
		this.SQConnection = conn
	}

	// Set configuation
	cfg := Config{
		Database: defaultDatabase,
	}
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	} else {
		this.schema = cfg.Database
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<eventqueue"
	if this.schema != "" {
		str += fmt.Sprintf(" schema=%q", this.schema)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "eventqueue"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	if err := this.createTables(ctx); err != nil {
		provider.Print(ctx, "failed to create tables: ", err)
		return err
	}

	// Emit incoming events
FOR_LOOP:
	for {
		select {
		case evt := <-this.C:
			for name, s := range this.S {
				select {
				case s <- evt:
					// no-op
				default:
					provider.Print(ctx, "eventqueue: cannot send on blocked channel: ", name)
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
		go func() {
			this.indexEvent(ctx, evt)
		}()
		break
	default:
		panic("eventqueue: cannot post on blocked channel")
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
