package nginx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	// Package imports

	"github.com/hashicorp/go-multierror"
	"github.com/mutablelogic/go-server/pkg/event"
	task "github.com/mutablelogic/go-server/pkg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type t struct {
	task.Task
	*exec.Cmd
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*t, error) {
	this := new(t)
	this.Cmd = exec.Command(p.path, p.flags...)
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *t) String() string {
	str := "<nginx"
	if t.Cmd != nil {
		str += fmt.Sprintf(" cmd=%q", t.Cmd.Path)
		str += fmt.Sprintf(" args=%q", t.Cmd.Args)
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *t) Run(ctx context.Context) error {
	var result error
	var wg sync.WaitGroup

	// Create a cancelable context
	child, cancel := context.WithCancel(ctx)

	// Run nginx in the background, cancel when an error is returned
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := t.exec(child, t.Cmd); err != nil {
			result = multierror.Append(result, err)
			cancel()
		}
	}()

	// Wait for context to be cancelled
	<-child.Done()

	// Stop nginx server if not already stopped
	if !t.Cmd.ProcessState.Exited() {
		t.Emit(event.Infof(ctx, EventStop, "Stopping nginx"))
		if err := t.Signal(ctx, "quit"); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Wait for nginx to exit
	wg.Wait()

	// Return any errors
	return result
}

func (t *t) Signal(ctx context.Context, cond string) error {
	flags := []string{"-s", cond}
	for _, flag := range t.Cmd.Args[1:] {
		if flag != "-s" {
			flags = append(flags, flag)
		}
	}
	return t.exec(ctx, exec.Command(t.Cmd.Path, flags...))
}

// Execute the command in the foreground, and emit events
// for stdout and stderr
func (t *t) exec(ctx context.Context, cmd *exec.Cmd) error {
	// Pipes for reading stdout and stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Read from stdout and stderr in the background
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		t.logger(ctx, EventInfo, stdout)
	}()
	go func() {
		defer wg.Done()
		t.logger(ctx, EventError, stderr)
	}()

	// Wait for stdout and stderr to be closed
	wg.Wait()

	// Wait for command to exit, return any errors
	return cmd.Wait()
}

////////////////////////////////////////////////////////////////////////////////
// PROCESS LOG FILES

func (t *t) logger(ctx context.Context, key Event, fh io.Reader) {
	buf := bufio.NewReader(fh)
	for {
		if line, err := buf.ReadBytes('\n'); err == io.EOF {
			return
		} else if err != nil {
			fmt.Println("ERROR", err)
			t.Emit(event.Error(ctx, err))
		} else {
			fmt.Println("LOG", key, strings.TrimSpace(string(line)))
		}
	}
}
