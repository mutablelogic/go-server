package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	// Modules
	. "github.com/djthorpe/go-server"
	"github.com/rjeczalik/notify"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Indexer struct {
	name string
	path string
	c    chan<- IndexerEvent
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 100
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(name, path string, c chan<- IndexerEvent) (*Indexer, error) {
	this := new(Indexer)

	// Check path argument
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, ErrBadParameter.With("invalid path: ", strconv.Quote(path))
	} else {
		this.name = name
		this.path = path
		this.c = c
	}

	// Return success
	return this, nil
}

func (this *Indexer) Run(ctx context.Context) error {
	ch := make(chan notify.EventInfo, defaultCapacity)
	if err := notify.Watch(filepath.Join(this.path, "..."), ch, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil {
		return err
	}

FOR_LOOP:
	for {
		// Dispatch events to index files and folders until context is cancelled
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-ch:
			info, _ := os.Stat(evt.Path())
			switch evt.Event() {
			case notify.Create:
				this.ProcessPath(EVENT_TYPE_ADDED, evt.Path(), info)
			case notify.Remove:
				this.ProcessPath(EVENT_TYPE_REMOVED, evt.Path(), info)
			case notify.Rename:
				this.ProcessPath(EVENT_TYPE_RENAMED, evt.Path(), info)
			case notify.Write:
				this.ProcessPath(EVENT_TYPE_CHANGED, evt.Path(), info)
			}
		}
	}

	// Stop notify and close channel
	notify.Stop(ch)
	close(ch)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Indexer) String() string {
	str := "<indexer"
	if this.name != "" {
		str += fmt.Sprintf(" name=%q", this.name)
	}
	if this.path != "" {
		str += fmt.Sprintf(" path=%q", this.path)
	}
	return str + ">"
}
