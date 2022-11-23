package log

import (
	"context"
	"log"
	"os"
	"sync"

	// Package imports
	ctx "github.com/mutablelogic/go-server/pkg/context"
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
	this.Logger = log.New(os.Stderr, p.Label(), p.flags)
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Print(parent context.Context, v ...interface{}) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	prefix := ctx.NameLabel(parent)
	if prefix != "." {
		t.Logger.SetPrefix(prefix + ": ")
	} else {
		t.Logger.SetPrefix("")
	}
	t.Logger.Print(v...)
}

func (t *t) Printf(parent context.Context, f string, v ...interface{}) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	prefix := ctx.NameLabel(parent)
	if prefix != "." {
		t.Logger.SetPrefix(prefix + ": ")
	} else {
		t.Logger.SetPrefix("")
	}
	t.Logger.Printf(f, v...)
}
