package log

import (
	// Package imports
	"context"
	"log"
	"os"
	"sync"

	task "github.com/mutablelogic/go-server/pkg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	sync.Mutex
	task.Task
	*log.Logger
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	this := new(t)
	this.Logger = log.New(os.Stderr, p.Prefix(), p.flags)
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Print(ctx context.Context, v ...interface{}) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	t.Logger.Print(v...)
}

func (t *t) Printf(ctx context.Context, f string, v ...interface{}) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	t.Logger.Printf(f, v...)
}
