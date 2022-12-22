package log

import (
	"context"
	"fmt"
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
// STRINGIFY

func (t *t) String() string {
	str := "<log"
	if w := t.Logger.Writer(); w != nil {
		if f, ok := w.(*os.File); ok {
			str += fmt.Sprintf(" writer=%q", f.Name())
		}
	}
	if t.Logger.Flags()&log.Ldate != 0 {
		str += " date"
	}
	if t.Logger.Flags()&log.Ltime != 0 {
		str += " time"
	}
	if t.Logger.Flags()&log.Lmsgprefix != 0 {
		str += " msgprefix"
	}
	if t.Logger.Flags()&log.Llongfile != 0 {
		str += " longfile"
	}
	if t.Logger.Flags()&log.Lshortfile != 0 {
		str += " shortfile"
	}
	if t.Logger.Flags()&log.LstdFlags != 0 {
		str += " stdflags"
	}
	if t.Logger.Flags()&log.Lmicroseconds != 0 {
		str += " microseconds"
	}
	if t.Logger.Flags()&log.LUTC != 0 {
		str += " utc"
	}
	return str + ">"
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
