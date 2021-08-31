package main

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	indexer "github.com/djthorpe/go-server/pkg/indexer"
	sq "github.com/djthorpe/go-sqlite"
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Database string   `yaml:"database"`
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Exclude  []string `yaml:"exclude"`
}

type plugin struct {
	sq.SQConnection
	*indexer.Indexer
	*indexer.Database

	// Channel used for recieving indexer events
	c chan indexer.IndexerEvent
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultDatabase = "main"
	defaultCapacity = 100
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the indexer plugin
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Get configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	} else if cfg.Name == "" || cfg.Path == "" {
		provider.Print(ctx, "GetConfig: missing name or path")
		return nil
	} else if cfg.Database == "" {
		cfg.Database = defaultDatabase
	}

	// Get sqlite, make database
	if conn, ok := provider.GetPlugin(ctx, "sqlite").(sq.SQConnection); !ok {
		provider.Print(ctx, "missing sqlite dependency")
		return nil
	} else if db, err := indexer.NewDatabase(conn, cfg.Database); err != nil {
		provider.Print(ctx, "NewDatabase: ", err)
		return nil
	} else {
		this.Database = db
	}

	// Make channel for indexing events
	this.c = make(chan indexer.IndexerEvent, defaultCapacity)

	// Make path absolute, make indexer
	if abspath, err := filepath.Abs(cfg.Path); err != nil {
		provider.Print(ctx, "NewIndexer: ", err)
		return nil
	} else if indexer, err := indexer.New(cfg.Name, abspath, this.c); err != nil {
		provider.Print(ctx, "NewIndexer: ", err)
		return nil
	} else {
		this.Indexer = indexer
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<indexer"
	if this.Indexer != nil {
		str += " " + this.Indexer.String()
	}
	if this.Database != nil {
		str += " " + this.Database.String()
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "indexer"
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	var result error
	var wg sync.WaitGroup

	// Process indexer events in background
	wg.Add(1)
	go func() {
		defer wg.Done()
	FOR_LOOP:
		for {
			select {
			case <-ctx.Done():
				break FOR_LOOP
			case evt := <-this.c:
				switch evt.Type() {
				case indexer.EVENT_TYPE_ADDED, indexer.EVENT_TYPE_CHANGED:
					if id, err := this.Database.AddFile(evt.Key(), evt.Path(), evt.FileInfo()); err != nil {
						provider.Print(ctx, "IndexerEvent: ", err)
					} else {
						provider.Print(ctx, "IndexerEvent: ", id)
					}
				case indexer.EVENT_TYPE_RENAMED:
					provider.Print(ctx, "IndexerEvent: ", evt)
				case indexer.EVENT_TYPE_REMOVED:
					provider.Print(ctx, "IndexerEvent: ", evt)
				}
			}
		}
	}()

	// Index filesystem in foreground after a short delay
	<-time.After(time.Second * 3)
	if err := this.Indexer.Walk(ctx); err != nil {
		result = multierror.Append(result, err)
	}

	// Run indexer in background once filesystem has been indexed
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := this.Indexer.Run(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Wait for all go-routines to finish
	wg.Wait()

	// Close channel, release resources
	close(this.c)

	// Return success
	return result
}
